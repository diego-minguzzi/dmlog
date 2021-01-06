# dmlog
## Overview
A Go package to log the program execution.
The logging is asynchronous.

The user can add one or more log sinks where log messages are added, according to the configured format.
The current version supports:
-  console sink
-  file sink
-  rolling file sink

## How to get the package
In order to add the package to your go environment, use the command:

`
go get github.com/diego-minguzzi/dmlog
`

## Usage example

```
package main

import "dmlog"
import "log"

func myFunc() {
    defer dmlog.MethodStartEnd()()
}

func main() {
    // Releases all log resources when the program terminates.
    defer dmlog.Terminate()
	
    // Adds a sink that writes to the console.
    consSinkId, err := dmlog.AddConsoleSink( dmlog.InfoSeverity)
    if err!=nil {
        log.Panicln("AddConsoleSink() failed:",err)
    }
    // Sets a custom format for the console sink.
    if ! dmlog.SetSinkOutputFormat(consSinkId, dmlog.LogMessageType, 
                                 dmlog.FilenameLineFmt,dmlog.LineEndFmt,dmlog.SeverityFmt,dmlog.TextFmt,dmlog.LineEndFmt) {
        log.Panicln("SetSinkOutputFormat() failed.")
    }
    
    // Adds a sink that writes to a file, using the default format.
    _, err= dmlog.AddFileSinkCreate( "log_file.txt", dmlog.DebugSeverity)
    if err!=nil {
        log.Panicln("AddFileSinkCreate() failed:",err)
    }

    dmlog.Debug("Debug message")
    dmlog.Warn("Warning message")
    
    // By default print messages are printed without filename, line number, etc.
    dmlog.Print("A printed message") 
    myFunc()
    
    // [...]
}
```
On my Linux laptop, the console output is:
```
/home/minguzzi/go/src/my_usage.go:32 
[WRN] Warning message 

A printed message 

```
(The console has Info severity).

While the content of the log file log_file.txt is:
```
/home/minguzzi/go/src/my_usage.go:31 2021-01-06 22:29:40.834 
[DBG] Debug message 

/home/minguzzi/go/src/my_usage.go:32 2021-01-06 22:29:40.834 
[WRN] Warning message 

A printed message 

/home/minguzzi/go/src/my_usage.go:7 2021-01-06 22:29:40.834 
[DBG] main.myFunc() started 

/home/minguzzi/go/src/my_usage.go:7 2021-01-06 22:29:40.834 
[DBG] main.myFunc() terminated 
```

## Documentation
[Docs hosted by GitHub](https://godoc.org/github.com/diego-minguzzi/dmlog)

## Todo
* Add a Panic() mathod, similar to log.Panic()

