package dmlog

import "fmt"
import "io/ioutil"
import "os"
import "testing"
import "time"

//--------------------------------------------------------------------------------------------------
func TestFileSinkCreate( t *testing.T) {
    tmpFile, err := ioutil.TempFile("","log_file_sink_test_")
    if err!=nil {
        t.Error(t.Name(),`TempFile() failed:`,err)            
        return
    }    
    tmpFilename := tmpFile.Name()        
    tmpFile.Close()    
    os.Remove(tmpFilename)

    filename := tmpFilename+".txt"
    sinkId, err := AddFileSink( filename, false, DebugSeverity, true)  
    if err!=nil {
        t.Error(t.Name(),`AddFileSink() failed:`,err)            
        return
    }
    Debug("sinkId:",sinkId)
    defer os.Remove(filename)
    SetSeverity( DebugSeverity)
    Debug("Debug message")
    Info("Info message")
    time.Sleep(200*time.Millisecond)

    fileInfo, err := os.Stat(filename)
    if err!=nil {
        t.Error(t.Name(),`os.Stat() failed: got:`,err)            
        return
    }
    if fileInfo.Size()<=0 {
        t.Error(t.Name(),`Unexpected empty output file:`,filename)            
    }    
    time.Sleep(100*time.Millisecond)
    ClearSinks()
}

//--------------------------------------------------------------------------------------------------
func TestFileSinkAppend( t *testing.T) {
    tmpFile, err := ioutil.TempFile("","log_file_sink_test_")
    if err!=nil {
        t.Error(t.Name(),`TempFile() failed:`,err)            
        return
    }    
    filename := tmpFile.Name()
    defer os.Remove(filename)
    
    err = tmpFile.Close()    
    if err!=nil {
    t.Error(t.Name(),`Close() failed:`,err)            
        return
    }

    sinkId, err := AddFileSink( filename, true, DebugSeverity, true)  
    if err!=nil {
        t.Error(t.Name(),`AddFileSink() failed:`,err)            
        return
    }
    SetSeverity( DebugSeverity)
    var got, want bool
    got = Debug("Debug message")
    want = true
    if got!=want {
        t.Error(t.Name(),`Debug() failed: got:`,got,`want:`,want)            
    }

    SetSeverity( WarningSeverity)
    got = Info("Info message")
    want = false
    if got!=want {
        t.Error(t.Name(),`Info() failed: got:`,got,`want:`,want)            
    }

    SetMessageSinkSeverity( sinkId, InfoSeverity)
    time.Sleep(200*time.Millisecond)

    fileInfo, err := os.Stat(filename)
    if err!=nil {
        t.Error(t.Name(),`os.Stat() failed: got:`,err)            
        return
    }
    if fileInfo.Size()<=0 {
        t.Error(t.Name(),`Unexpected empty output file:`,filename)            
    }    

    SetSeverity( DebugSeverity)
    time.Sleep(100*time.Millisecond)
    ClearSinks()
}

/* Example of how to log to a file that is overwritten at every execution. */
func ExampleAddFileSinkCreate() {
    _, err := AddFileSinkCreate( "log.txt", DebugSeverity) 
    if err != nil {
        fmt.Fprintln(os.Stderr,"AddFileSinkCreate() failed:",err)
        return
    }
    
    Debug("Debug message")
    time.Sleep(100*time.Millisecond)
    ClearSinks()    
}

/* Example of how to log to a file that appends to an existing file or it creates a new file if
 * it does not exist. */
func ExampleAddFileSinkAppend() {
    _, err := AddFileSinkAppend( "log.txt", DebugSeverity) 
    if err != nil {
        fmt.Fprintln(os.Stderr,"AddFileSinkAppend() failed:",err)
        return
    }
    
    Debug("Debug message")  
}
