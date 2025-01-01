//go:build darwin

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#import <Cocoa/Cocoa.h>
int
SetActivationPolicy(void) {
    [NSApp setActivationPolicy:NSApplicationActivationPolicyAccessory];
    return 0;
}
*/
import "C"
import (
	"fmt"
	"time"
)

func setActivationPolicy() {
	fmt.Println("Setting ActivationPolicy")
	C.SetActivationPolicy()
}

func hideMeIfNotWindows() {
	thisApp.Lifecycle().SetOnStarted(func() {
		go func() {
			time.Sleep(200 * time.Millisecond)
			setActivationPolicy()
		}()
	})
}
