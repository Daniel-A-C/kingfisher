//go:build darwin

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#import <Cocoa/Cocoa.h>
#import <dlfcn.h>

extern void kfOnWindowSummoned(char *source);

typedef int CGSConnectionID;
typedef CGSConnectionID (*kfCGSDefaultConnFn)(void);
typedef OSStatus (*kfCGSSetBlurRadiusFn)(CGSConnectionID cid, NSInteger wid, int radius);

static BOOL kfReopenObserverInstalled = NO;
static NSString *kfLastExternalBundleID = nil;
static CFAbsoluteTime kfLastDeactivateTime = 0;

static void kfApplyWindowBlur(double blur) {
    dispatch_async(dispatch_get_main_queue(), ^{
        double clamped = blur < 0.0 ? 0.0 : (blur > 1.0 ? 1.0 : blur);
        int radius = (int)(clamped * 40.0 + 0.5);

        kfCGSDefaultConnFn defaultConn = (kfCGSDefaultConnFn)dlsym(RTLD_DEFAULT, "CGSDefaultConnectionForThread");
        kfCGSSetBlurRadiusFn setBlur = (kfCGSSetBlurRadiusFn)dlsym(RTLD_DEFAULT, "CGSSetWindowBackgroundBlurRadius");
        BOOL hasCGSBlur = (defaultConn != NULL && setBlur != NULL);

        for (NSWindow *win in [NSApp windows]) {
            [win setOpaque:NO];
            [win setBackgroundColor:[NSColor clearColor]];

            NSView *content = [win contentView];
            for (NSView *sub in [content subviews]) {
                if ([sub isKindOfClass:[NSVisualEffectView class]]) {
                    if (hasCGSBlur) {
                        // Hide macOS's murky NSVisualEffectView material so opacity scales
                        // cleanly from 0% to 100% and blur is controlled purely by radius.
                        [sub setHidden:YES];
                    } else {
                        [sub setHidden:(clamped <= 0.001)];
                        [sub setAlphaValue:clamped];
                    }
                }
            }

            if (hasCGSBlur) {
                setBlur(defaultConn(), [win windowNumber], radius);
            }
        }
    });
}

static void kfRestoreMainWindow(void) {
    for (NSWindow *win in [NSApp windows]) {
        if ([win isMiniaturized]) {
            [win deminiaturize:nil];
        }
        if (![win isVisible]) {
            [win makeKeyAndOrderFront:nil];
        }
    }
}

static const char *kfDetectSummonSource(void) {
    CFAbsoluteTime now = CFAbsoluteTimeGetCurrent();
    if (kfLastExternalBundleID != nil && (now - kfLastDeactivateTime) < 3.0) {
        NSString *lower = [kfLastExternalBundleID lowercaseString];
        if ([lower containsString:@"raycast"]) {
            return "raycast";
        }
        if ([lower containsString:@"spotlight"]) {
            return "spotlight";
        }
    }
    return "activate";
}

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

        if (!kfReopenObserverInstalled) {
            kfReopenObserverInstalled = YES;

            NSNotificationCenter *wsNC = [[NSWorkspace sharedWorkspace] notificationCenter];
            [wsNC addObserverForName:NSWorkspaceDidDeactivateApplicationNotification
                              object:nil
                               queue:[NSOperationQueue mainQueue]
                          usingBlock:^(NSNotification *note) {
                              NSRunningApplication *app = note.userInfo[NSWorkspaceApplicationKey];
                              NSString *bid = app.bundleIdentifier;
                              if (bid && ![bid isEqualToString:[[NSBundle mainBundle] bundleIdentifier]]) {
                                  kfLastExternalBundleID = [bid copy];
                                  kfLastDeactivateTime = CFAbsoluteTimeGetCurrent();
                              }
                          }];

            NSNotificationCenter *nc = [NSNotificationCenter defaultCenter];
            [nc addObserverForName:NSApplicationDidBecomeActiveNotification
                            object:nil
                             queue:[NSOperationQueue mainQueue]
                        usingBlock:^(__unused NSNotification *note) {
                            kfRestoreMainWindow();
                            kfOnWindowSummoned((char *)kfDetectSummonSource());
                        }];
            [nc addObserverForName:NSApplicationDidUnhideNotification
                            object:nil
                             queue:[NSOperationQueue mainQueue]
                        usingBlock:^(__unused NSNotification *note) {
                            kfRestoreMainWindow();
                        }];
        }
    });
}
*/
import "C"

import (
	"sync"
)

var (
	summonMu       sync.RWMutex
	summonCallback func(source string)
)

func setSummonCallback(fn func(source string)) {
	summonMu.Lock()
	summonCallback = fn
	summonMu.Unlock()
}

//export kfOnWindowSummoned
func kfOnWindowSummoned(source *C.char) {
	src := "activate"
	if source != nil {
		src = C.GoString(source)
	}
	summonMu.RLock()
	fn := summonCallback
	summonMu.RUnlock()
	if fn != nil {
		go fn(src)
	}
}

// hideNativeWindowButtons hides macOS's built-in traffic lights while keeping
// NSWindowStyleMaskTitled (and therefore native macOS rounded window corners),
// ensures activating or unhiding Kingfisher always restores its window, and
// notifies the Go app whenever Kingfisher is summoned (including from Raycast).
func hideNativeWindowButtons() {
	C.kfHideNativeWindowButtons()
}

// setNativeWindowBlur sets the behind-window Gaussian blur intensity (0.0–1.0)
// independently of window opacity, hiding macOS's murky NSVisualEffectView material.
func setNativeWindowBlur(blur float64) {
	C.kfApplyWindowBlur(C.double(blur))
}
