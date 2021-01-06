package dmlog

import "fmt"

/* Issues a debug log message.
   In all functions issuing a log message with signature like 
   func Debug(v ...interface{}) bool 
   function arguments are printed using the default formats. 
   Spaces are added between operands when neither is a string. */
func Debug(v ...interface{}) bool { 
    return addLogMessage( fmt.Sprint(v...), DebugSeverity, LogMessageType, nil, defaultSkip)
}

// Issues a warning message.
func Warn(v ...interface{}) bool { 
    return addLogMessage( fmt.Sprint(v...), WarningSeverity, LogMessageType, nil, defaultSkip)
}

// Issues an info message.
func Info(v ...interface{}) bool { 
    return addLogMessage( fmt.Sprint(v...), InfoSeverity, LogMessageType, nil, defaultSkip)
}

// Prints a log message.
func Print(v ...interface{}) bool { 
    return addLogMessage( fmt.Sprint(v...), PrintSeverity, PrintMessageType, nil, defaultSkip)
}

func LogPrint(v ...interface{}) bool { 
    return addLogMessage( fmt.Sprint(v...), PrintSeverity, PrintMessageType, nil, defaultSkip)
}

// Issues a message with error severity level.
func Error(v ...interface{}) bool { 
    return addLogMessage( fmt.Sprint(v...), ErrorSeverity, LogMessageType, nil, defaultSkip)
}

// Issues a message with fatal severity level.
func Fatal(v ...interface{}) bool { 
    return addLogMessage( fmt.Sprint(v...), FatalSeverity, LogMessageType, nil, defaultSkip)
}

// Logs the execution of a method.
func MethodExecuted() bool {
    var caller callerDetails
    getCallerDetails( &caller, defaultSkip)
    return addLogMessage( caller.funcName+"() executed", DebugSeverity, LogMessageType, &caller, defaultSkip)
}

// Logs a method when it starts and terminates.  The returned function must be deferred.
func MethodStartEnd() func() {
    var caller callerDetails
    getCallerDetails( &caller, defaultSkip)
    addLogMessage( caller.funcName+ "() started", DebugSeverity, LogMessageType, &caller, defaultSkip)
    return func() {
        addLogMessage( caller.funcName+"() terminated", DebugSeverity, LogMessageType, &caller, defaultSkip+1)
    }
}


