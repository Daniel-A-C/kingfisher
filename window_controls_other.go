//go:build !darwin

package main

func setSummonCallback(_ func(source string)) {}

func hideNativeWindowButtons() {}
