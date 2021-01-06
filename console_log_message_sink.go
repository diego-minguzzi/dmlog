package dmlog

import "fmt"
import "os"

// Implementation of a log sink that prints messages on the console.
type consoleLogMessageSink struct {
    BaseLogMessageSink
}

/* Adds a log message sink that prints on the console.
   In case error is nil, the returned message sink id can be used later to modify the severity
   threshold.
 */
func AddConsoleSink( threshold LogSeverity) (MessageSinkId, error) {
    msgSink := newConsoleLogMessageSink( threshold, false)
    return addMessageSink( msgSink)
}

//--------------------------------------------------------------------------------------------------
func newConsoleLogMessageSink( threshold LogSeverity, isFrequentFlush bool) *consoleLogMessageSink {
    messageTypeToFormat := map[MessageType]LogFormatItems {
        LogMessageType: defaultLogFormat(),
        PrintMessageType: defaultPrintFormat(),
    }
    obj := consoleLogMessageSink{ BaseLogMessageSink  {
                                    threshold:threshold,                                                       
                                    isFrequentFlush:isFrequentFlush,
                                    messageTypeToFormat: messageTypeToFormat,} } 
    return &obj
}

//--------------------------------------------------------------------------------------------------
func (c *consoleLogMessageSink) SetSeverity( threshold LogSeverity) {
    c.threshold= threshold
}
    
//--------------------------------------------------------------------------------------------------
func (c *consoleLogMessageSink) Severity() LogSeverity {
    return c.threshold
}   
    
//--------------------------------------------------------------------------------------------------
func (c *consoleLogMessageSink) OnLogMessage( msg *LogMessage) {
    if msg.severity.IsGreaterOrEqualThan( c.threshold) {
        format, ok := c.messageTypeToFormat[ msg.messageType ]
        if ok {
            fmt.Print( formatLogMessage( msg, &format) )
        } else {
            fmt.Println( msg.text)
        }
        if c.isFrequentFlush {
            c.flush()
        }
    }
}
    
//--------------------------------------------------------------------------------------------------
func (c *consoleLogMessageSink) flush() {
    os.Stdout.Sync()    
}
    
//--------------------------------------------------------------------------------------------------
func (c *consoleLogMessageSink) SetFlush( isFrequentFlush bool) {
    c.isFrequentFlush= isFrequentFlush
}

//--------------------------------------------------------------------------------------------------
func (c *consoleLogMessageSink) setSinkFormat( messageType MessageType, 
                                                format LogFormatItems) bool {
    return c.BaseLogMessageSink.setSinkFormat( messageType, format)
}
    
//--------------------------------------------------------------------------------------------------
func (c *consoleLogMessageSink) terminate() {
    c.flush()
}

