//go:build !darwin

package main

func setSummonCallback(_ func(source string)) {}

func hideNativeWindowButtons() {}

func setNativeWindowBlur(_ float64) {}
