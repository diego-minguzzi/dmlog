// Package implementing a logging facility.
package dmlog

import "log"
import "sync"
import "time"

const defaultSinksCapacity int = 5
const defaultCapChLogMessages int = 100
const defaultSkip int = 1

const fatalLogTerminated string = "Log facility already terminated."

// Unique identifier of a Message sing
type MessageSinkId int

// Each message type can be given a specific format.
type MessageType int8

// The supported message types.
const (
    /* Log messages, usually for debugging purpose. */
    LogMessageType MessageType = iota
    /* Print messages consumed also by the user. */
    PrintMessageType
)

// How much alarming is a given log.
type LogSeverity int8

// The supported severity levels, in order of increasing severity.
const (
    DebugSeverity LogSeverity = iota
    InfoSeverity
    PrintSeverity
    WarningSeverity
    ErrorSeverity
    FatalSeverity
) 

type LogMessage struct {
    text        string
    severity    LogSeverity
    messageType MessageType
    timestamp   time.Time 
    filename    string
    line        int
    funcName    string
}

//--------------------------------------------------------------------------------------------------
type cmdSetMessageSinkThreshold struct {
    sinkId      MessageSinkId
    threshold   LogSeverity
}

//--------------------------------------------------------------------------------------------------
// Implements the Stringable interface
func (t LogSeverity) String() string {
    switch t {
        case DebugSeverity:   return "DBG"
        case InfoSeverity:    return "INF"
        case PrintSeverity:   return "PRN"
        case WarningSeverity: return "WRN"
        case ErrorSeverity:   return "ERR"
        case FatalSeverity:   return "FAT"
    }
    return "Unknown"
}

type LogMessageSink interface {
    SetSeverity( threshold LogSeverity)
    
    Severity() LogSeverity    
    
    OnLogMessage( msg *LogMessage)
    
    flush()

    setSinkFormat( messageType MessageType, format LogFormatItems) bool
    
    SetFlush( isFrequentFlush bool)
    
    /* The message sink is terminated: it must release its resources. */
    terminate()    
}

//--------------------------------------------------------------------------------------------------
var context struct {
    /* The severity is atomic, it is checked before a message is sent. */
    severity LogSeverity
    
    // Mutex to access the severity field.
    mtxSeverity sync.RWMutex
        
    /* The most important channel, where log messages are sent to the dispatcher.*/
    chLogMessages chan LogMessage

    /* The channel to send requests to the message dispatcher and receive the corresponding reply.*/
    chRequest chan interface{}

    chReply chan interface{}
    
    /* When the message is closed, the tracing must be terminated.*/
    chReqTerminate chan struct{}    
    chReplyTerminate chan struct{}  
}

type BaseLogMessageSink struct {
    threshold LogSeverity
    isFrequentFlush bool
    messageTypeToFormat map[MessageType]LogFormatItems
}

func (b *BaseLogMessageSink)setSinkFormat( messageType MessageType, format LogFormatItems) bool {
    b.messageTypeToFormat[messageType]= format
    return true
} 

//--------------------------------------------------------------------------------------------------
func init() {
    context.severity= DebugSeverity
    context.chRequest= make(chan interface{})
    context.chReply= make(chan interface{})
    context.chLogMessages= make( chan LogMessage, defaultCapChLogMessages)
    context.chReqTerminate= make( chan struct{}) 
    context.chReplyTerminate= make( chan struct{}) 
    go messageDispatcher()
}

// Determines whether the tracing facility was terminated.
func IsTerminated() bool {
    select {
        case <- context.chReqTerminate:
            return true
        default:
            return false
    }
}

/* Terminate the tracing service.  After termination, all calls to the methods will result in a 
 * fatal.*/
func Terminate() {
    if ! IsTerminated() {    
        close( context.chReqTerminate)     
        <- context.chReplyTerminate
    }
}

/* Sets the global severity threshold.  
 * Messages below the threshold are not forwarded to the sinks. */
func SetSeverity( severity LogSeverity){
    if IsTerminated() {
      log.Panic( fatalLogTerminated)
    }
    
    context.mtxSeverity.Lock()
    defer context.mtxSeverity.Unlock()
    
    if (severity != context.severity){
        context.severity= severity
    }    
}

// Retrieves the global severity threshold.
func Severity() LogSeverity {
    if IsTerminated() {
      log.Panic( fatalLogTerminated)
    }
    
    context.mtxSeverity.RLock()
    defer context.mtxSeverity.RUnlock()
    return context.severity
} 

/* Sets the severity of the given sink.*/
func SetMessageSinkSeverity( sinkId MessageSinkId, threshold LogSeverity) bool {
    return reqMessageSinkThreshold( sinkId, threshold)
}

/* Set the format of for a message type of a given sink.
 * The sink is identified by a valid sinkId.
 * formatItems is a sequence of LogFormatItem elements. */
func SetSinkOutputFormat( sinkId MessageSinkId, 
                          messageType MessageType, 
                          formatItems ...LogFormatItem) bool {
    return reqSetSinkFormat( sinkId, messageType, formatItems...)
}

/* Terminate and remove all current sinks. */
func ClearSinks() bool {
    return reqClearSinks() 
}

//--------------------------------------------------------------------------------------------------
func addLogMessage( text string, 
                    severity LogSeverity, 
                    messageType MessageType, 
                    forcedCaller *callerDetails, 
                    skip int ) bool {
    context.mtxSeverity.RLock()
    defer context.mtxSeverity.RUnlock()

    if severity.IsGreaterOrEqualThan(context.severity) && (! IsTerminated()) {
        var message LogMessage
        message.text= text
        message.severity= severity
        message.messageType= messageType
        message.timestamp= time.Now()

        if nil!=forcedCaller {
            message.funcName= forcedCaller.funcName
            message.filename= forcedCaller.filename
            message.line = forcedCaller.line
        } else {
            var caller callerDetails
            _ = getCallerDetails( &caller, /*skip*/skip+1)
            message.funcName= caller.funcName
            message.filename= caller.filename
            message.line = caller.line
        }

        context.chLogMessages <- message   
        return true        
    }
    return false
}

//--------------------------------------------------------------------------------------------------
func (s LogSeverity) IsGreaterOrEqualThan(that LogSeverity) bool {
    switch s {
        case DebugSeverity: {
            switch that {
                case DebugSeverity: return true                
                case InfoSeverity:  return false
                case PrintSeverity:  return false
                case WarningSeverity: return false
                case ErrorSeverity: return false
                case FatalSeverity: return false
            }
        }        
        case InfoSeverity: {
            switch that {
                case DebugSeverity:   return true                
                case InfoSeverity:    return true
                case PrintSeverity:   return false                
                case WarningSeverity: return false                
                case ErrorSeverity:   return false                
                case FatalSeverity:   return false
            }
        }
        case PrintSeverity: {
            switch that {
                case DebugSeverity:   return true                
                case InfoSeverity:    return true
                case PrintSeverity:   return true                
                case WarningSeverity: return false                
                case ErrorSeverity:   return false                
                case FatalSeverity:   return false
            }
        }
        case WarningSeverity: {
            switch that {
                case DebugSeverity:   return true
                case InfoSeverity:    return true
                case PrintSeverity:   return true
                case WarningSeverity: return true
                case ErrorSeverity:   return false  
                case FatalSeverity:   return false
            }
        }
        case ErrorSeverity: {
            switch that {
                case DebugSeverity:   return true                       
                case InfoSeverity:    return true                   
                case PrintSeverity:   return true
                case WarningSeverity: return true 
                case ErrorSeverity:   return true         
                case FatalSeverity:   return false
            }
        }
        case FatalSeverity: return true
    } 
    return false
}

//--------------------------------------------------------------------------------------------------
func addMessageSink(messageSink LogMessageSink) (MessageSinkId,error) {
    if messageSink == nil {
        panic("addMessageSink(): invalid argument")
    }
    return reqMessageSink( &messageSink)
}
