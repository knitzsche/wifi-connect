// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2017 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package main

import (
	"fmt"
	"testing"
)

type mock0 struct{}

func (mock *mock0) SnapGet(s string) (string, error) {
	return "snapgetreturn", nil
}

type Config struct {
	AString string
	ABool   bool
}

func TestSnapGetStr0(t *testing.T) {
	c := GetTestClient(&mock0{})
	config := &Config{}
	config.AString = "defaultstring"
	c.snapGetStr("key", &config.AString)
	if config.AString != "snapgetreturn" {
		t.Errorf("snapGetStr expected snapgetreturn but got %s", config.AString)
	}
}

type mock1 struct{}

func (mock *mock1) SnapGet(s string) (string, error) {
	return "", fmt.Errorf("intentional error 1")
}

func TestSnapGetStr1(t *testing.T) {
	c := GetTestClient(&mock1{})
	config := &Config{}
	config.AString = "defaultstring"
	c.snapGetStr("key", &config.AString)
	if config.AString != "defaultstring" {
		t.Errorf("snapGetStr expected defaultstring but got %s", config.AString)
	}
}

type mock2 struct{}

func (mock *mock2) SnapGet(s string) (string, error) {
	return "true", nil
}

func TestSnapGetBool0(t *testing.T) {
	c := GetTestClient(&mock2{})
	config := &Config{}
	config.ABool = false
	c.snapGetBool("key", &config.ABool)
	if !config.ABool {
		t.Errorf("snapGetBool should be true but is %t", config.ABool)
	}
}

type mock3 struct{}

func (mock *mock3) SnapGet(s string) (string, error) {
	return "false", nil
}

func TestSnapGetBool1(t *testing.T) {
	c := GetTestClient(&mock3{})
	config := &Config{}
	config.ABool = true
	c.snapGetBool("key", &config.ABool)
	if config.ABool {
		t.Errorf("snapGetBool should be false but is %t", config.ABool)
	}
}

type mock4 struct{}

func (mock *mock4) SnapGet(s string) (string, error) {
	return "", nil
}

func TestSnapGetBool2(t *testing.T) {
	c := GetTestClient(&mock4{})
	config := &Config{}
	config.ABool = true
	c.snapGetBool("key", &config.ABool)
	if !config.ABool {
		t.Errorf("snapGetBool should be true but is %t", config.ABool)
	}
}
