package main

import "path"
import "encoding/xml"
import "encoding/json"
import "strings"
import "log"
import "os"
import "os/exec"
import "io"
import "flag"

/*import "bytes"*/

import "pkg.linuxdeepin.com/lib/dbus"

//Supports target now
const (
	QML    string = "qml"
	GoLang        = "golang"
	PyQt          = "pyqt"
)

func GetInterfaceInfo(ifc _Interface) dbus.InterfaceInfo {
	inFile := path.Join(INFOS.Config.InputDir, ifc.XMLFile)
	if _, err := os.Stat(inFile); err == nil {
		reader, err := os.Open(inFile)
		if err != nil {
			panic(err.Error() + "(File:" + inFile + ")")
		}
		decoder := xml.NewDecoder(reader)
		obj := dbus.NodeInfo{}
		decoder.Decode(&obj)
		for _, ifcInfo := range obj.Interfaces {
			if ifcInfo.Name == ifc.Interface {
				return ifc.handlleBlackList(ifcInfo)
			}
		}
		reader.Close()
	}
	panic("Not Found Interface " + ifc.Interface)
}

type _Interface struct {
	OutFile, XMLFile, Dest, ObjectPath, Interface, ObjectName, TestPath string
	BlackMethods                                                        []string
	BlackProperties                                                     []string
	BlackSignals                                                        []string
}

func (ifc _Interface) handlleBlackList(data dbus.InterfaceInfo) dbus.InterfaceInfo {
	for _, name := range ifc.BlackMethods {
		var methods []dbus.MethodInfo
		for _, mInfo := range data.Methods {
			if mInfo.Name != name {
				methods = append(methods, mInfo)
			}
		}
		data.Methods = methods
	}
	for _, name := range ifc.BlackProperties {
		var properties []dbus.PropertyInfo
		for _, pInfo := range data.Properties {
			if pInfo.Name != name {
				properties = append(properties, pInfo)
			}
		}
		data.Properties = properties
	}
	for _, name := range ifc.BlackSignals {
		var signals []dbus.SignalInfo
		for _, sInfo := range data.Signals {
			if sInfo.Name != name {
				signals = append(signals, sInfo)
			}
		}
		data.Signals = signals
	}
	return data
}

type _Config struct {
	Target       string
	NotExportBus bool
	OutputDir    string
	InputDir     string
	PkgName      string
	DestName     string
	BusType      string
}

type Infos struct {
	Interfaces []_Interface
	Config     _Config
	outputs    map[string]io.Writer
}

var INFOS *Infos

func parseInfo() {
	if INFOS != nil {
		panic("Don't call multime the function of parseInfo")
	}
	INFOS = new(Infos)
	var outputPath, inputFile, target string
	flag.StringVar(&outputPath, "out", "out", "the directory to save the generated code")
	flag.StringVar(&inputFile, "in", "dbus.in.json", "the config file path")
	flag.StringVar(&target, "target", "", "now support QML/PyQt/GoLang target")
	flag.Parse()

	f, err := os.Open(inputFile)
	if err != nil {
		panic(err)
	}
	dec := json.NewDecoder(f)
	err = dec.Decode(INFOS)
	if err != nil {
		panic(err)
	}

	if outputPath != "out" {
		INFOS.Config.OutputDir = outputPath
	} else if len(INFOS.Config.OutputDir) == 0 {
		INFOS.Config.OutputDir = outputPath
	}
	if target != "" {
		INFOS.Config.Target = target
	}

	INFOS.outputs = make(map[string]io.Writer)
	os.MkdirAll(INFOS.Config.OutputDir, 0755)
	busType := strings.ToLower(INFOS.Config.BusType)
	INFOS.Config.BusType = busType
	if busType != "session" && busType != "system" {
		log.Fatal("Didn't support bus type", busType)
	}

	// convert target
	INFOS.Config.Target = strings.ToLower(INFOS.Config.Target)
	if INFOS.Config.PkgName == "" {
		INFOS.Config.PkgName = getMember(INFOS.Config.DestName)
		if INFOS.Config.PkgName == "" {
			log.Fatal("Didn't specify an PkgName and can't calclus an valid PkgName by DestName:" + INFOS.Config.DestName)
		}
	}
	if INFOS.Config.Target == GoLang {
		INFOS.Config.BusType = upper(INFOS.Config.BusType)
		for _, ifc := range INFOS.Interfaces {
			if INFOS.outputs[ifc.OutFile], err = os.Create(path.Join(INFOS.Config.OutputDir, ifc.OutFile+".go")); err != nil {
				panic(err)
			}
			renderInterfaceInit(INFOS.outputs[ifc.OutFile])
		}
	} else if INFOS.Config.Target == PyQt {
		INFOS.Config.BusType = lower(INFOS.Config.BusType)
		for _, ifc := range INFOS.Interfaces {
			if INFOS.outputs[ifc.OutFile], err = os.Create(path.Join(INFOS.Config.OutputDir, ifc.OutFile+".py")); err != nil {
				panic(err)
			}
			renderInterfaceInit(INFOS.outputs[ifc.OutFile])
		}
	} else if INFOS.Config.Target == QML {
		INFOS.Config.BusType = lower(INFOS.Config.BusType)
		for _, ifc := range INFOS.Interfaces {
			if INFOS.outputs[ifc.OutFile], err = os.Create(path.Join(INFOS.Config.OutputDir, ifc.OutFile+".h")); err != nil {
				panic(err)
			}
			renderInterfaceInit(INFOS.outputs[ifc.OutFile])
		}
		renderQMLProject()
	} else {
		log.Fatal(`Didn't supported target , please set Target to "Golang" or "PyQt" or "QML"`)
	}
}

func main() {
	parseInfo()
	var writer io.Writer
	var err error
	if INFOS.Config.Target == GoLang {
		if writer, err = os.Create(path.Join(INFOS.Config.OutputDir, "init.go")); err != nil {
			panic(err)
		}
	} else if INFOS.Config.Target == PyQt {
		if writer, err = os.Create(path.Join(INFOS.Config.OutputDir, "__init__.py")); err != nil {
			panic(err)
		}
	} else if INFOS.Config.Target == QML {
		if writer, err = os.Create(path.Join(INFOS.Config.OutputDir, "plugin.h")); err != nil {
			panic(err)
		}
	}
	renderMain(writer)
	writer.(*os.File).Close()

	defer func() {
		exec.Command("gofmt", "-w", INFOS.Config.OutputDir).Start()
		for _, w := range INFOS.outputs {
			w.(*os.File).Close()
		}
		if INFOS.Config.Target == QML {
			testQML()
		}
	}()
	for _, ifc := range INFOS.Interfaces {
		writer = INFOS.outputs[ifc.OutFile]

		inFile := path.Join(INFOS.Config.InputDir, ifc.XMLFile)
		if _, err := os.Stat(inFile); err == nil {
			info := GetInterfaceInfo(ifc)
			renderInterface(info, writer, ifc.Interface, ifc.ObjectName)
			/*if ifc.TestPath != "" {*/
			/*var test_writer io.Writer*/
			/*test_writer, err = os.Create(path.Join(INFOS.Config.OutputDir, path.Base(ifc.OutFile)+"_test.go"))*/
			/*render(ifc.TestPath, INFOS.Config.PkgName, ifc.ObjectName, test_writer, info)*/
			/*}*/
		} else {
			panic(inFile + " dind't exists")
			conn, _ := dbus.SystemBus()
			var xml string
			if err := conn.Object(ifc.Dest, dbus.ObjectPath(ifc.ObjectPath)).Call("org.freedesktop.DBus.Introspectable.Introspect", 0).Store(&xml); err != nil {
				panic(err.Error() + "Interface " + ifc.Interface + " is can't dynamic introspect")
			}
			renderInterface(GetInterfaceInfo(ifc), writer, ifc.Interface, ifc.ObjectName)

		}
	}
}
