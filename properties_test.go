/*
 * This file is part of PropertiesMap library.
 *
 * Copyright 2017 Arduino AG (http://www.arduino.cc/)
 *
 * PropertiesMap library is free software; you can redistribute it and/or modify
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
	"path/filepath"
	"runtime"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPropertiesBoardsTxt(t *testing.T) {
	p, err := Load(filepath.Join("testdata", "boards.txt"))

	require.NoError(t, err)

	require.Equal(t, "Processor", p.Get("menu.cpu"))
	require.Equal(t, "32256", p.Get("ethernet.upload.maximum_size"))
	require.Equal(t, "{build.usb_flags}", p.Get("robotMotor.build.extra_flags"))

	ethernet := p.SubTree("ethernet")
	require.Equal(t, "Arduino Ethernet", ethernet.Get("name"))
}

func TestPropertiesTestTxt(t *testing.T) {
	p, err := Load(filepath.Join("testdata", "test.txt"))

	require.NoError(t, err)

	require.Equal(t, 4, p.Size())
	require.Equal(t, "value = 1", p.Get("key"))

	switch value := runtime.GOOS; value {
	case "linux":
		require.Equal(t, "is linux", p.Get("which.os"))
	case "windows":
		require.Equal(t, "is windows", p.Get("which.os"))
	case "darwin":
		require.Equal(t, "is macosx", p.Get("which.os"))
	default:
		require.FailNow(t, "unsupported OS")
	}
}

func TestExpandPropsInString(t *testing.T) {
	aMap := NewMap()
	aMap.Set("key1", "42")
	aMap.Set("key2", "{key1}")

	str := "{key1} == {key2} == true"

	str = aMap.ExpandPropsInString(str)
	require.Equal(t, "42 == 42 == true", str)
}

func TestExpandPropsInString2(t *testing.T) {
	p := NewMap()
	p.Set("key2", "{key2}")
	p.Set("key1", "42")

	str := "{key1} == {key2} == true"

	str = p.ExpandPropsInString(str)
	require.Equal(t, "42 == {key2} == true", str)
}

func TestDeleteUnexpandedPropsFromString(t *testing.T) {
	p := NewMap()
	p.Set("key1", "42")
	p.Set("key2", "{key1}")

	str := "{key1} == {key2} == {key3} == true"

	str = p.ExpandPropsInString(str)
	str = DeleteUnexpandedPropsFromString(str)
	require.Equal(t, "42 == 42 ==  == true", str)
}

func TestDeleteUnexpandedPropsFromString2(t *testing.T) {
	p := NewMap()
	p.Set("key2", "42")

	str := "{key1} == {key2} == {key3} == true"

	str = p.ExpandPropsInString(str)
	str = DeleteUnexpandedPropsFromString(str)
	require.Equal(t, " == 42 ==  == true", str)
}

func TestPropertiesRedBeearLabBoardsTxt(t *testing.T) {
	p, err := Load(filepath.Join("testdata", "redbearlab_boards.txt"))

	require.NoError(t, err)

	require.Equal(t, 83, p.Size())
	require.Equal(t, "Blend", p.Get("blend.name"))
	require.Equal(t, "arduino:arduino", p.Get("blend.build.core"))
	require.Equal(t, "0x2404", p.Get("blendmicro16.pid.0"))

	ethernet := p.SubTree("blend")
	require.Equal(t, "arduino:arduino", ethernet.Get("build.core"))
}

func TestSubTreeForMultipleDots(t *testing.T) {
	p := NewMap()
	p.Set("root.lev1.prop", "hi")
	p.Set("root.lev1.prop2", "how")
	p.Set("root.lev1.prop3", "are")
	p.Set("root.lev1.prop4", "you")
	p.Set("root.lev1", "A")

	lev1 := p.SubTree("root.lev1")
	require.Equal(t, "you", lev1.Get("prop4"))
	require.Equal(t, "hi", lev1.Get("prop"))
	require.Equal(t, "how", lev1.Get("prop2"))
	require.Equal(t, "are", lev1.Get("prop3"))
}

func TestPropertiesBroken(t *testing.T) {
	_, err := Load(filepath.Join("testdata", "broken.txt"))

	require.Error(t, err)
}

func TestGetSetBoolean(t *testing.T) {
	m := NewMap()
	m.Set("a", "true")
	m.Set("b", "false")
	m.Set("c", "hello")
	m.SetBoolean("e", true)
	m.SetBoolean("f", false)

	require.True(t, m.GetBoolean("a"))
	require.False(t, m.GetBoolean("b"))
	require.False(t, m.GetBoolean("c"))
	require.False(t, m.GetBoolean("d"))
	require.True(t, m.GetBoolean("e"))
	require.False(t, m.GetBoolean("f"))
	require.Equal(t, "true", m.Get("e"))
	require.Equal(t, "false", m.Get("f"))
}

func TestKeysMethod(t *testing.T) {
	m := NewMap()
	m.Set("k1", "value")
	m.Set("k2", "othervalue")
	m.Set("k3.k4", "anothevalue")
	m.Set("k5", "value")

	k := m.Keys()
	sort.Strings(k)
	require.Equal(t, "[k1 k2 k3.k4 k5]", fmt.Sprintf("%s", k))

	v := m.Values()
	sort.Strings(v)
	require.Equal(t, "[anothevalue othervalue value value]", fmt.Sprintf("%s", v))
}
