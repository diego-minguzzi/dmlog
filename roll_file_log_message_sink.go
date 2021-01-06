package dmlog

import "fmt"
import "log"
import "os"
import "path"
import "path/filepath"
import "sort"
import "strings"
import "time"

const defaultCapChStrLog int = 100

// The extension of the log file
const fileExtension string = ".txt"

const timestampSeparator string = "_"

type Bytes  uint64
type KBytes uint64
const kBytesToBytes Bytes = 1024

// Implementation of a log sink that prints messages to a set of rolling files.
type rollFileLogMessageSink struct {
    BaseLogMessageSink
    
    dirPath           string
    filePrefix        string
    numMaxFiles       int
    maxFileSize       Bytes
    currFile          *os.File
    currFileSize      int

    chStrLog          chan string
    chReqTerminate    chan struct{} 
    chReplyTerminate  chan struct{} 
}

/* Adds a trace message sink that write messages into log files stored in the directory dirPath.
   The created files have the given filePrefix.
   At most numMaxFiles matching the dirPath and the filePrefix are kept, then the oldest is erased.
   Moreover, each log file size is at most maxFileSize, expressed into kBytes.
   In case error is nil, the returned message sink id can be used later to modify the severity 
   threshold.*/
func AddRollFileSink( dirPath         string,
                      filePrefix      string,
                      numMaxFiles     int,
                      maxFileSize     KBytes, 
                      threshold       LogSeverity) (MessageSinkId, error) {
    msgSink, err := newRollFileLogMessageSink(dirPath, filePrefix, numMaxFiles, maxFileSize, threshold)
    if err!=nil {
        return MessageSinkId(0), err
    } 
    return addMessageSink( msgSink)
}

type fileInfoSliceType []os.FileInfo

func (f fileInfoSliceType) Len() int {
    return len(f)
}

func (f fileInfoSliceType) Less(i, j int) bool {
    return f[i].ModTime().Before(f[j].ModTime())
}

func (f fileInfoSliceType) Swap(i, j int) {
    f[i], f[j] = f[j], f[i]
}

//--------------------------------------------------------------------------------------------------
func newRollFileLogMessageSink( dirPath       string,
                                filePrefix    string,
                                numMaxFiles   int,
                                maxFileSize   KBytes, 
                                threshold     LogSeverity) (*rollFileLogMessageSink, error) {
    // Evaluates dirPath
    dirPathInfo, err := os.Stat( dirPath)
    if err!=nil {
        return nil,fmt.Errorf("os.Stat() failed on %s",dirPath)
    }
    if !dirPathInfo.IsDir() {
        return nil,fmt.Errorf("dir path %s is not a directory",dirPath)
    }
    // Evaluates the file prefix
    var trimmedFilePrefix = strings.TrimSpace(filePrefix)
    if len(trimmedFilePrefix)<=0 {
        return nil,fmt.Errorf("invalid file prefix %s",trimmedFilePrefix)
    }

    if numMaxFiles<=0 {
        return nil,fmt.Errorf("invalid numMaxFiles parameter %d",numMaxFiles)
    }

    if maxFileSize<=0 {
        return nil,fmt.Errorf("invalid maxFileSize parameter %d",maxFileSize)
    }
        
    logFile, err := createNewRollFile( trimmedFilePrefix, dirPath,  numMaxFiles) 
    if err!=nil {
        return nil,err
    }

    messageTypeToFormat := map[MessageType]LogFormatItems {
        LogMessageType:   defaultLogFormat(),
        PrintMessageType: defaultPrintFormat(),
    }

    var result = rollFileLogMessageSink {
                        BaseLogMessageSink: BaseLogMessageSink{
                            threshold:threshold,
                            isFrequentFlush: false,
                            messageTypeToFormat: messageTypeToFormat,
                        },  
                        dirPath: dirPath,
                        filePrefix: trimmedFilePrefix,
                        numMaxFiles: numMaxFiles,
                        maxFileSize: Bytes( maxFileSize)*kBytesToBytes,
                        currFile: logFile,
                        currFileSize:0,
                        chStrLog: make( chan string, defaultCapChStrLog),
                        chReqTerminate: make( chan struct{}),
                        chReplyTerminate: make( chan struct{}),
                    }
    go rollFileSinkHandler( &result)
    return &result,nil
}

//--------------------------------------------------------------------------------------------------
func (c *rollFileLogMessageSink) SetSeverity( threshold LogSeverity) {
    c.threshold= threshold
}
    
//--------------------------------------------------------------------------------------------------
func (c *rollFileLogMessageSink) Severity() LogSeverity {
    return c.threshold
}   
    
//--------------------------------------------------------------------------------------------------
func (r *rollFileLogMessageSink) OnLogMessage( msg *LogMessage) {
    if msg.severity.IsGreaterOrEqualThan( r.threshold) {
        format, ok := r.messageTypeToFormat[ msg.messageType ]
        if ok {
            r.chStrLog <- formatLogMessage( msg, &format)
        } else {
            r.chStrLog <- msg.text
        }
    }
}
    
//--------------------------------------------------------------------------------------------------
func (r *rollFileLogMessageSink) flush() {}
    
//--------------------------------------------------------------------------------------------------
func (r *rollFileLogMessageSink) SetFlush( isFrequentFlush bool) {
    r.isFrequentFlush= isFrequentFlush
}

//--------------------------------------------------------------------------------------------------
func (r *rollFileLogMessageSink) setSinkFormat( messageType MessageType, 
                                                  format LogFormatItems) bool {
    return r.BaseLogMessageSink.setSinkFormat( messageType, format)
}
    
//--------------------------------------------------------------------------------------------------
func (r *rollFileLogMessageSink) terminate() {
    close( r.chReqTerminate)
    <- r.chReplyTerminate 
}

//--------------------------------------------------------------------------------------------------
// Retrieves the number of files in the directory dirPath, that match the fileGlob pattern.
func numFilesMatching(dirPath string, fileGlob string) (int, error) {
    filePattern := path.Join( dirPath, fileGlob)
    matchFiles, err := filepath.Glob( filePattern)
    if err!=nil {
        return 0, fmt.Errorf("Glob() failed:%s",err)
    }
    return len(matchFiles),nil
}

//--------------------------------------------------------------------------------------------------
/* Deletes the numOlderFiles oldest files matching the pattern numOlderFiles in the 
 * directory dirPath */
func deleteOlderFiles( dirPath string, fileGlob string, numOlderFiles int ) error {
    filePattern := filepath.Join( dirPath, fileGlob)
    matchFiles, err := filepath.Glob( filePattern)
    if err!=nil {
        return fmt.Errorf("filepath.Glob() failed:%s",err)
    }
    numMatchingFiles := len(matchFiles)
    fileInfoSlice := make( []os.FileInfo, 0, numMatchingFiles) 
    for _, matchFile := range matchFiles {
        fileInfo, err := os.Stat( matchFile)
        if err!=nil {
            return fmt.Errorf("os.Stat() failed on %s",matchFile)
        } else {
            fileInfoSlice= append( fileInfoSlice, fileInfo)
        }
    }
    if numMatchingFiles<numOlderFiles {
        log.Println(numMatchingFiles,"files: no file to delete.")
    } else {
        sort.Sort(sort.Reverse( fileInfoSliceType(fileInfoSlice) ))
        for indx:=0; indx<numOlderFiles; indx++ {
          filename :=  fileInfoSlice[indx].Name()
          err := os.Remove( filepath.Join( dirPath, fileInfoSlice[indx].Name()) )
          if err!=nil {
              return fmt.Errorf("failed while trying to remove the file %s:%s",filename,err)
          } 
        }
    }

    return nil 
}
//--------------------------------------------------------------------------------------------------
func createNewRollFile( filePrefix string, dirPath string, numMaxFiles int) (*os.File,error) {

    logFilesGlob := filePrefix+ "*"
    numFiles, err := numFilesMatching( dirPath, logFilesGlob)
    if err!=nil {
        return nil,fmt.Errorf("numFilesMatching() failed: %s",err)
    }
    if numFiles >= numMaxFiles {
        err = deleteOlderFiles( dirPath, logFilesGlob, numFiles - numMaxFiles + 1)
        return nil,err
    }

    return createNewFile( dirPath, filePrefix, fileExtension)
}

//--------------------------------------------------------------------------------------------------
func createNewFile( dirPath string, filePrefix string, fileExtension string) (*os.File,error) {
    var filename = filePrefix+ 
                   timestampSeparator+ 
                   timestampString( time.Now(), timestampSeparator)+
                   fileExtension 
    var filepath = filepath.Join(dirPath, filename)
    return os.Create( filepath)
}

//--------------------------------------------------------------------------------------------------
func rollFileSinkHandler( ctx *rollFileLogMessageSink) {
    for terminate:= false; !terminate; {
        select {
            case strLog := <- ctx.chStrLog: {
                rollFileSinkOnNewStrLog( ctx, strLog)                
            }
            case <- ctx.chReqTerminate: {
                for hasLogStrings:= true; hasLogStrings; {
                    select {
                        case strLog := <- ctx.chStrLog: {
                            rollFileSinkOnNewStrLog( ctx, strLog)                
                        }
                        default: 
                            hasLogStrings = false                      
                    }                       
                }
                if nil!=ctx.currFile {
                    ctx.currFile.Close()
                    ctx.currFile = nil 
                }
                terminate= true                
            }
        }
    }
    close( ctx.chReplyTerminate)
}

//--------------------------------------------------------------------------------------------------
func rollFileSinkOnNewStrLog( ctx *rollFileLogMessageSink, strMessage string) {
    var strMessageLen = len(strMessage)
    if (ctx.currFile != nil ) &&
       ( Bytes(ctx.currFileSize + strMessageLen) < ctx.maxFileSize ) {
        fmt.Fprint( ctx.currFile, strMessage)
        ctx.currFileSize += strMessageLen
    } else {
        if ctx.currFile != nil {
            ctx.currFile.Close()
        }
        newFile, err := createNewRollFile( ctx.filePrefix, ctx.dirPath, ctx.numMaxFiles) 
        if err==nil {
            ctx.currFile = newFile
            fmt.Fprint( ctx.currFile, strMessage)
            ctx.currFileSize = strMessageLen         
        }                    
    }             
}
