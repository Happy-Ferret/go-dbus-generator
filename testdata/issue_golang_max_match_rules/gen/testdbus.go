/*This file is automatically generated by pkg.deepin.io/dbus-generator. Don't edit it*/
package testdbus

import "pkg.deepin.io/lib/dbus"
import "pkg.deepin.io/lib/dbus/property"
import "reflect"
import "sync"
import "runtime"
import "fmt"
import "errors"

/*prevent compile error*/
var _ = fmt.Println
var _ = runtime.SetFinalizer
var _ = sync.NewCond
var _ = reflect.TypeOf
var _ = property.BaseObserver{}

type TestDBus struct {
	Path     dbus.ObjectPath
	DestName string
	core     *dbus.Object

	signals       map[<-chan *dbus.Signal]struct{}
	signalsLocker sync.Mutex

	TimerProp *dbusPropertyTestDBusTimerProp
}

func (obj *TestDBus) _createSignalChan() <-chan *dbus.Signal {
	obj.signalsLocker.Lock()
	ch := getBus().Signal()
	obj.signals[ch] = struct{}{}
	obj.signalsLocker.Unlock()
	return ch
}
func (obj *TestDBus) _deleteSignalChan(ch <-chan *dbus.Signal) {
	obj.signalsLocker.Lock()
	delete(obj.signals, ch)
	getBus().DetachSignal(ch)
	obj.signalsLocker.Unlock()
}
func DestroyTestDBus(obj *TestDBus) {
	obj.signalsLocker.Lock()
	defer obj.signalsLocker.Unlock()
	if obj.signals == nil {
		return
	}
	for ch, _ := range obj.signals {
		getBus().DetachSignal(ch)
	}
	obj.signals = nil

	runtime.SetFinalizer(obj, nil)

	dbusRemoveMatch("type='signal',path='" + string(obj.Path) + "',interface='org.freedesktop.DBus.Properties',sender='" + obj.DestName + "',member='PropertiesChanged'")
	dbusRemoveMatch("type='signal',path='" + string(obj.Path) + "',interface='com.deepin.TestDBus',sender='" + obj.DestName + "',member='PropertiesChanged'")

	dbusRemoveMatch("type='signal',path='" + string(obj.Path) + "',interface='com.deepin.TestDBus',sender='" + obj.DestName + "',member='TimerSignal'")

	obj.TimerProp.Reset()
}

func (obj *TestDBus) ConnectTimerSignal(callback func(arg0 string)) func() {
	sigChan := obj._createSignalChan()
	go func() {
		for v := range sigChan {
			if v.Path != obj.Path || v.Name != "com.deepin.TestDBus.TimerSignal" || 1 != len(v.Body) {
				continue
			}
			if reflect.TypeOf(v.Body[0]) != reflect.TypeOf((*string)(nil)).Elem() {
				continue
			}

			callback(v.Body[0].(string))
		}
	}()
	return func() {
		obj._deleteSignalChan(sigChan)
	}
}

type dbusPropertyTestDBusTimerProp struct {
	*property.BaseObserver
	core *dbus.Object
}

func (this *dbusPropertyTestDBusTimerProp) SetValue(notwritable interface{}) {
	fmt.Println("com.deepin.TestDBus.TimerProp is not writable")
}

func (this *dbusPropertyTestDBusTimerProp) Get() string {
	return this.GetValue().(string)
}
func (this *dbusPropertyTestDBusTimerProp) GetValue() interface{} /*string*/ {
	var r dbus.Variant
	err := this.core.Call("org.freedesktop.DBus.Properties.Get", 0, "com.deepin.TestDBus", "TimerProp").Store(&r)
	if err == nil && r.Signature().String() == "s" {
		return r.Value().(string)
	} else {
		fmt.Println("dbusProperty:TimerProp error:", err, "at com.deepin.TestDBus")
		return *new(string)
	}
}
func (this *dbusPropertyTestDBusTimerProp) GetType() reflect.Type {
	return reflect.TypeOf((*string)(nil)).Elem()
}

func NewTestDBus(destName string, path dbus.ObjectPath) (*TestDBus, error) {
	if !path.IsValid() {
		return nil, errors.New("The path of '" + string(path) + "' is invalid.")
	}

	core := getBus().Object(destName, path)

	obj := &TestDBus{Path: path, DestName: destName, core: core, signals: make(map[<-chan *dbus.Signal]struct{})}

	obj.TimerProp = &dbusPropertyTestDBusTimerProp{&property.BaseObserver{}, core}

	dbusAddMatch("type='signal',path='" + string(path) + "',interface='org.freedesktop.DBus.Properties',sender='" + destName + "',member='PropertiesChanged'")
	dbusAddMatch("type='signal',path='" + string(path) + "',interface='com.deepin.TestDBus',sender='" + destName + "',member='PropertiesChanged'")
	sigChan := obj._createSignalChan()
	go func() {
		typeString := reflect.TypeOf("")
		typeKeyValues := reflect.TypeOf(map[string]dbus.Variant{})
		typeArrayValues := reflect.TypeOf([]string{})
		for v := range sigChan {
			if v.Name == "org.freedesktop.DBus.Properties.PropertiesChanged" &&
				len(v.Body) == 3 &&
				reflect.TypeOf(v.Body[0]) == typeString &&
				reflect.TypeOf(v.Body[1]) == typeKeyValues &&
				reflect.TypeOf(v.Body[2]) == typeArrayValues &&
				v.Body[0].(string) == "com.deepin.TestDBus" {
				props := v.Body[1].(map[string]dbus.Variant)
				for key, _ := range props {
					if false {
					} else if key == "TimerProp" {
						obj.TimerProp.Notify()
					}
				}
			} else if v.Name == "com.deepin.TestDBus.PropertiesChanged" && len(v.Body) == 1 && reflect.TypeOf(v.Body[0]) == typeKeyValues {
				for key, _ := range v.Body[0].(map[string]dbus.Variant) {
					if false {
					} else if key == "TimerProp" {
						obj.TimerProp.Notify()
					}
				}
			}
		}
	}()

	dbusAddMatch("type='signal',path='" + string(obj.Path) + "',interface='com.deepin.TestDBus',sender='" + obj.DestName + "',member='TimerSignal'")

	runtime.SetFinalizer(obj, func(_obj *TestDBus) { DestroyTestDBus(_obj) })
	return obj, nil
}
