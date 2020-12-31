package dmlog

import "io/ioutil"
import "log"
import "path"
import "path/filepath"
import "os"
import "testing"
import "time"

func TestRollFileSink( t *testing.T) {
    tempDirName, err := ioutil.TempDir("", "roll_file_sink_test")
    if err != nil {
        t.Error(t.Name(),`TempDir() failed:`,err)            
        return
    }
    log.Println("tempDirName:",tempDirName)
    filePrefix := "roll_file"
    const numMaxFiles = 3
    const maxFileSize = KBytes(100)
    sinkId, err := AddRollFileSink( tempDirName,filePrefix,numMaxFiles,maxFileSize,DebugSeverity)
    defer os.RemoveAll( tempDirName)
    if err != nil {
        t.Error(t.Name(),`AddRollFileSink() failed:`,err)            
        return
    }
    
    log.Println("sindId:",sinkId)
    Debug("Debug message ",time.Now())
    Info("Info message ",time.Now())
    Print("Print message ",time.Now())

    for indx:=0; indx<100; indx++ {
        for j:=0; j<10; j++ {
            Debug("Roll sink test:",time.Now()," message #", indx+j+1,".................................")
        }
        time.Sleep( 1*time.Millisecond)
    }
    time.Sleep( 1*time.Second)
    
    ClearSinks()

    writtenFiles, err := filepath.Glob( path.Join(tempDirName, filePrefix+"*") )
    if err != nil {
        t.Error(t.Name(),`filepath.Glob() failed:`,err)            
        return
    }
    gotNumWrittenFiles := len(writtenFiles)   
    if gotNumWrittenFiles!=numMaxFiles {
        t.Error(t.Name(),`got`,gotNumWrittenFiles,`written files. Want:`,numMaxFiles)            
    }
}

