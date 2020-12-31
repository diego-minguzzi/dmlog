package dmlog

import "fmt"
import "os"

// Implementation of a log sink that prints messages to a single file.
type fileLogMessageSink struct {
    BaseLogMessageSink
    outFile *os.File
}

/*------------------------------------------------------------------------------------------------*/
/* Adds a log message sink that write messages to the specified file.
 * If a file with the same name already exists, its content is erased.
 * In case error is nil, the returned message sink id can be used later to modify the severity 
 * threshold.*/
func AddFileSinkCreate( filename string, threshold LogSeverity) (MessageSinkId, error) {
    return AddFileSink( filename, false, threshold, false)
}

/*------------------------------------------------------------------------------------------------*/
/* Adds a log message sink that append messages to the specified file.
 * In case error is nil, the returned message sink id can be used later to modify the severity 
 * threshold.*/
func AddFileSinkAppend( filename string, threshold LogSeverity) (MessageSinkId, error) {
    return AddFileSink( filename, true, threshold, false)
}

/*------------------------------------------------------------------------------------------------*/
/* Adds a log message sink that prints on the specified file.
 * In case error is nil, the returned message sink id can be used later to modify the severity 
 * threshold.   */
func AddFileSink( filename string, 
                  appendExisting bool, 
                  threshold LogSeverity, 
                  isFrequentFlush bool) (MessageSinkId, error) {
    msgSink, err := newFileLogMessageSink( filename, appendExisting, threshold, isFrequentFlush)
    if err!=nil {
        return 0, err
    }
    return addMessageSink( msgSink)
}

//--------------------------------------------------------------------------------------------------
func newFileLogMessageSink( filename string,
                            appendExisting bool, 
                            threshold LogSeverity, 
                            isFrequentFlush bool) (*fileLogMessageSink, error) {
    messageTypeToFormat := map[MessageType]LogFormatItems {
        LogMessageType: defaultLogFormat(),
        PrintMessageType: defaultPrintFormat(),
    }
    var file *os.File
    var err error
    if appendExisting {
        file, err = os.OpenFile( filename, os.O_CREATE | os.O_APPEND | os.O_WRONLY, 0) 
        if err != nil {
            return nil, fmt.Errorf("failed while trying to append to the file %s:%s",filename,err)
        }        
    } else {
        file, err = os.Create( filename)
        if err != nil {
            return nil, fmt.Errorf("failed while trying to create the file %s:%s",filename,err)
        }        
    }

    obj := fileLogMessageSink{ BaseLogMessageSink: BaseLogMessageSink {
                                    threshold:threshold,                                                       
                                    isFrequentFlush:isFrequentFlush,
                                    messageTypeToFormat:messageTypeToFormat, },
                                 outFile: file, } 
    return &obj,nil
}

//--------------------------------------------------------------------------------------------------
func (f *fileLogMessageSink) SetSeverity( threshold LogSeverity) {
    f.threshold= threshold
}
    
//--------------------------------------------------------------------------------------------------
func (f *fileLogMessageSink) Severity() LogSeverity {
    return f.threshold
}   
    
//--------------------------------------------------------------------------------------------------
func (f *fileLogMessageSink) OnLogMessage( msg *LogMessage) {
    if msg.severity.IsGreaterOrEqualThan( f.threshold) {
        format, ok := f.messageTypeToFormat[ msg.messageType ]
        if ok {
            fmt.Fprint( f.outFile, formatLogMessage( msg, &format) )
        } else {
            fmt.Fprintln( f.outFile, msg.text)
        }
        if f.isFrequentFlush {
            f.flush()
        }
    }
}
    
//--------------------------------------------------------------------------------------------------
func (f *fileLogMessageSink) flush() {
    if f.outFile != nil {
        f.outFile.Sync()
    }
}
    
//--------------------------------------------------------------------------------------------------
func (f *fileLogMessageSink) SetFlush( isFrequentFlush bool) {
    f.isFrequentFlush= isFrequentFlush
}

//--------------------------------------------------------------------------------------------------
func (f *fileLogMessageSink) setSinkFormat( messageType MessageType, 
                                            format LogFormatItems) bool {
    return f.BaseLogMessageSink.setSinkFormat( messageType, format)
}
    
//--------------------------------------------------------------------------------------------------
func (f *fileLogMessageSink) terminate() {
    if f.outFile != nil {
        f.outFile.Close()
        f.outFile = nil
    }
}  
