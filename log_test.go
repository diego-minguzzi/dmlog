package dmlog

import "fmt"
import "log"
import "os"
import "testing"
import "time"

func mySimpleFunction(){
    MethodExecuted() 
}

func myTracedFunction(){
    Debug("myTracedFunction() executed.")
    defer MethodStartEnd()()
}

//--------------------------------------------------------------------------------------------------
func TestSetSeverity( t *testing.T) {
    t.Log(t.Name(),`TestSetSeverity()`)            
    want := InfoSeverity
    SetSeverity( want)
    got := Severity()
    t.Log(t.Name(),`Severity(): got`,got,`want`,want)            
    if got != want {
        t.Error(t.Name(),`Severity() failed: got`,got,`want`,want)            
    }
    
    want = DebugSeverity
    SetSeverity( want)
    got = Severity() 
    t.Log(t.Name(),`Severity(): got`,got,`want`,want)            
    if got != want {
        t.Error(t.Name(),`Severity() failed: got`,got,`want`,want)            
    }        
}

//--------------------------------------------------------------------------------------------------
func TestSeverityComparison( t *testing.T){
    t.Log(t.Name(),`executed`)            
    var testCases = []struct {  
        op LogSeverity 
        op2 LogSeverity
        want bool 
    }{
        {DebugSeverity, DebugSeverity, true},
        {InfoSeverity, DebugSeverity, true},
        {PrintSeverity, DebugSeverity, true},
        {WarningSeverity, DebugSeverity, true},
        {ErrorSeverity, DebugSeverity, true},
        {FatalSeverity, DebugSeverity, true},
        {DebugSeverity, InfoSeverity, false},
        {InfoSeverity, InfoSeverity, true},
        {PrintSeverity, InfoSeverity, true},
        {WarningSeverity, InfoSeverity, true},
        {ErrorSeverity, InfoSeverity, true},
        {FatalSeverity, InfoSeverity, true},
        {DebugSeverity, PrintSeverity, false},
        {InfoSeverity, PrintSeverity, false},
        {PrintSeverity, PrintSeverity, true},
        {WarningSeverity, PrintSeverity, true},
        {ErrorSeverity, PrintSeverity, true},
        {FatalSeverity, PrintSeverity, true},
        {DebugSeverity, WarningSeverity, false},
        {InfoSeverity, WarningSeverity, false},
        {PrintSeverity, WarningSeverity, false},
        {WarningSeverity, WarningSeverity, true},
        {ErrorSeverity, WarningSeverity, true},
        {FatalSeverity, WarningSeverity, true},
        {DebugSeverity, ErrorSeverity, false},
        {InfoSeverity, ErrorSeverity, false},
        {PrintSeverity, ErrorSeverity, false},
        {WarningSeverity, ErrorSeverity, false},
        {ErrorSeverity, ErrorSeverity, true},
        {FatalSeverity, ErrorSeverity, true},
        {DebugSeverity, FatalSeverity, false},
        {InfoSeverity, FatalSeverity, false},
        {PrintSeverity, FatalSeverity, false},
        {WarningSeverity, FatalSeverity,false},
        {ErrorSeverity, FatalSeverity, false},
        {FatalSeverity, FatalSeverity, true},
    }
    for indx, testCase := range testCases {
        got := testCase.op.IsGreaterOrEqualThan( testCase.op2)
        if testCase.want != got {
            t.Error(t.Name(),`failed: on test case #`, indx,"(",testCase.op,testCase.op2,`) got:`,got,`want`,testCase.want)                        
        }
    }
}

//--------------------------------------------------------------------------------------------------
func TestConsole( t *testing.T) {
    t.Log(t.Name())            
    sinkId, err := AddConsoleSink( DebugSeverity) 
    if err!=nil {
        t.Error(t.Name(),`AddConsoleSink() failed:`,err)                          
        return 
    }

    if sinkId < 0 {
        t.Error(t.Name(),`unexpected sink id value. Got:`,sinkId)                           
    }
    ok := SetSinkOutputFormat(sinkId, 
                              PrintMessageType, 
                              ShortTimestampFmt, TextFmt)
    if !ok {
        t.Error(t.Name(),"SetSinkOutputFormat() failed:",err)
        return
    }
    
    Debug("Example of debug message")   

    LogPrint("sinkId:",sinkId)
    Debug("Debug message")
    Info("Info message") 
    Warn("Warning message")
    Error("Error message")
    Fatal("Fatal message")
    n := 1000
    LogPrint("Print message n:",n)
    mySimpleFunction()
    myTracedFunction()
    time.Sleep(100*time.Millisecond)
    ClearSinks()
}

func aFunction() {
    defer MethodStartEnd()
}

/* Example of how to use a console sink. */
func ExampleAddConsoleSink() {

    sinkId,err := AddConsoleSink( DebugSeverity )
    if err != nil {
        fmt.Fprintln(os.Stderr,"AddConsoleSink() failed:",err)
        return
    }

    ok := SetSinkOutputFormat(sinkId, 
                              PrintMessageType, 
                              FilenameLineFmt, LongTimestampFmt, LineEndFmt, TextFmt)
    if !ok {
        fmt.Fprintln(os.Stderr,"AddConsoleSink() failed:",err)
        return
    }
    
    Debug("Example of debug message")   
    LogPrint("Example of print message")    

    ok = SetMessageSinkSeverity( sinkId, WarningSeverity)
    if !ok {
        fmt.Fprintln(os.Stderr,"SetMessageSinkSeverity() failed")
        return
    }

    n := 1000
    Debug("n:",n)   

    aFunction()

    Terminate()
}

func ExampleSetSinkOutputFormat(){
    // Releases all log resources when the program terminates.
    defer Terminate()
	
    // Adds a sink that writes to the console.
    sinkId, err := AddConsoleSink( DebugSeverity)
    if err!=nil {
        log.Panicln("AddConsoleSink() failed:",err)
	}
    /* Sets a custom format for the console sink, composed of:
     * filename and line, newline
     * severity tag, log message, newline
     */
    if ! SetSinkOutputFormat(sinkId, LogMessageType, 
                              FilenameLineFmt, LineEndFmt, SeverityFmt, TextFmt, LineEndFmt) {
        log.Panicln("SetSinkOutputFormat() failed.")
   }

    Debug("Debug log message")
}

