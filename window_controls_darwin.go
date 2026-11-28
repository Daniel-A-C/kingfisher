//go:build darwin

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#import <Cocoa/Cocoa.h>

static void kfHideNativeWindowButtons(void) {
    dispatch_async(dispatch_get_main_queue(), ^{
        for (NSWindow *win in [NSApp windows]) {
            NSButton *closeBtn = [win standardWindowButton:NSWindowCloseButton];
            NSButton *minBtn   = [win standardWindowButton:NSWindowMiniaturizeButton];
            NSButton *zoomBtn  = [win standardWindowButton:NSWindowZoomButton];
            if (closeBtn) [closeBtn setHidden:YES];
            if (minBtn)   [minBtn setHidden:YES];
            if (zoomBtn)  [zoomBtn setHidden:YES];
        }
    });
}
*/
import "C"

// hideNativeWindowButtons hides macOS's built-in traffic lights while keeping
// NSWindowStyleMaskTitled (and therefore native macOS rounded window corners).
func hideNativeWindowButtons() {
	C.kfHideNativeWindowButtons()
}
