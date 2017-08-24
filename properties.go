/*
 * This file is part of PropertiesMap library.
 *
 * Copyright 2017 Arduino LLC (http://www.arduino.cc/)
 *
 * Arduino Builder is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * As a special exception, you may use this file as part of a free software
 * library without restriction.  Specifically, if other files instantiate
 * templates or use macros or inline functions from this file, or you compile
 * this file and link it with other files to produce an executable, this
 * file does not by itself cause the resulting executable to be covered by
 * the GNU General Public License.  This exception does not however
 * invalidate any other reasons why the executable file might be covered by
 * the GNU General Public License.
 */

package properties

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"regexp"
	"runtime"
	"strings"
)

// Map is a container of properties
type Map map[string]string

var osSuffix string

func init() {
	switch value := runtime.GOOS; value {
	case "linux", "freebsd", "windows":
		osSuffix = runtime.GOOS
	case "darwin":
		osSuffix = "macosx"
	default:
		panic("Unsupported OS")
	}
}

func Load(filepath string) (Map, error) {
	bytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("Error reading file: %s", err)
	}

	text := string(bytes)
	text = strings.Replace(text, "\r\n", "\n", -1)
	text = strings.Replace(text, "\r", "\n", -1)

	properties := make(Map)

	for lineNum, line := range strings.Split(text, "\n") {
		if err := properties.parseLine(line); err != nil {
			return nil, fmt.Errorf("Error reading file (%s:%d): %s", filepath, lineNum, err)
		}
	}

	return properties, nil
}

func LoadFromSlice(lines []string) (Map, error) {
	properties := make(Map)

	for lineNum, line := range lines {
		if err := properties.parseLine(line); err != nil {
			return nil, fmt.Errorf("Error reading from slice (index:%d): %s", lineNum, err)
		}
	}

	return properties, nil
}

func (m Map) parseLine(line string) error {
	line = strings.TrimSpace(line)

	// Skip empty lines or comments
	if len(line) == 0 || line[0] == '#' {
		return nil
	}

	lineParts := strings.SplitN(line, "=", 2)
	if len(lineParts) != 2 {
		return fmt.Errorf("Invalid line format, should be 'key=value'")
	}
	key := strings.TrimSpace(lineParts[0])
	value := strings.TrimSpace(lineParts[1])

	key = strings.Replace(key, "."+osSuffix, "", 1)
	m[key] = value

	return nil
}

func SafeLoad(filepath string) (Map, error) {
	_, err := os.Stat(filepath)
	if os.IsNotExist(err) {
		return make(Map), nil
	}

	properties, err := Load(filepath)
	if err != nil {
		return nil, err
	}
	return properties, nil
}

func (m Map) FirstLevelOf() map[string]Map {
	newMap := make(map[string]Map)
	for key, value := range m {
		if strings.Index(key, ".") == -1 {
			continue
		}
		keyParts := strings.SplitN(key, ".", 2)
		if newMap[keyParts[0]] == nil {
			newMap[keyParts[0]] = make(Map)
		}
		newMap[keyParts[0]][keyParts[1]] = value
	}
	return newMap
}

func (m Map) SubTree(key string) Map {
	return m.FirstLevelOf()[key]
}

func (m Map) ExpandPropsInString(str string) string {
	replaced := true
	for i := 0; i < 10 && replaced; i++ {
		replaced = false
		for key, value := range m {
			newStr := strings.Replace(str, "{"+key+"}", value, -1)
			replaced = replaced || str != newStr
			str = newStr
		}
	}
	return str
}

func (m Map) Merge(sources ...Map) Map {
	for _, source := range sources {
		for key, value := range source {
			m[key] = value
		}
	}
	return m
}

func (m Map) Clone() Map {
	clone := make(Map)
	clone.Merge(m)
	return clone
}

func (m Map) Equals(other Map) bool {
	return reflect.DeepEqual(m, other)
}

func MergeMapsOfProperties(target map[string]Map, sources ...map[string]Map) map[string]Map {
	for _, source := range sources {
		for key, value := range source {
			target[key] = value
		}
	}
	return target
}

func DeleteUnexpandedPropsFromString(str string) string {
	rxp := regexp.MustCompile("\\{.+?\\}")
	return rxp.ReplaceAllString(str, "")
}
