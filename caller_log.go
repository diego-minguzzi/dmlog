package dmlog

import "runtime"

// Describes the details needed to log the caller of a given method.
type callerDetails struct {
    filename string
    line int
    funcName string
}

/* Even if it fails, caller has meaningful values. */
func getCallerDetails( caller *callerDetails, skip int) bool {
    if caller==nil {
        panic("Invalid caller argument.")
    }
    var pc uintptr
    var ok bool
    pc, caller.filename, caller.line, ok = runtime.Caller(/*skip*/skip+1)
    if !ok {
        caller.filename= "N/A"
        caller.line= 0
        caller.funcName= "N/A"
        return false
    }
    ptrFunc := runtime.FuncForPC(pc)
    if ptrFunc == nil {
        caller.funcName= "N/A"
        return false
    }
    caller.funcName= ptrFunc.Name()
    return true
}
