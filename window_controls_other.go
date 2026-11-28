//go:build !darwin

package main

func setSummonCallback(_ func(source string)) {}

func hideNativeWindowButtons() {}

func setNativeWindowTransparency(_ bool) {}
