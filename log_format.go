package dmlog

import "fmt"
import "strings"
import "time"

// Support for the format of log messages, when printed.

type LogFormatItem uint8

const defaultFormattedLogMessageCapacity int = 256
const longTimestampFormat string = "Mon,02 Jan 2006,15:04:05.000"
const shortTimestampFormat string = "2006-01-02 15:04:05.000"

// The supported log message format items.
const (
    FilenameFmt LogFormatItem = iota  // The filename of the source file
    FilenameLineFmt // Filename and line number
    FunctNameFmt // The function name 
    LineFmt
    ShortTimestampFmt
    LongTimestampFmt
    SeverityFmt
    TextFmt
    LineEndFmt
)

// For each everity it caches the corresponding plain text.
var severityToText map[LogSeverity]string

type LogFormatItems []LogFormatItem

// Retrieves the default format for print messages.
func defaultPrintFormat() []LogFormatItem {
    return []LogFormatItem{TextFmt,LineEndFmt }
}

// Retrieves the default format for plain log messages.
func defaultLogFormat() []LogFormatItem {
    return []LogFormatItem{FilenameLineFmt,ShortTimestampFmt,LineEndFmt,SeverityFmt,TextFmt,LineEndFmt}
}

func init() {
    severityToText = createMapSeverityToText()
}

func surroundWith( str string, left string, right string) string { return left+ str+ right }

// Create a map mapping each severity to its corresponding formatted text string.
func createMapSeverityToText() map[LogSeverity]string {
    result := make(map[LogSeverity]string)
    result[DebugSeverity]= surroundWith( DebugSeverity.String(),"[","]")
    result[InfoSeverity]= surroundWith( InfoSeverity.String(),"[","]")
    result[PrintSeverity]= surroundWith( PrintSeverity.String(),"[","]")
    result[WarningSeverity]= surroundWith( WarningSeverity.String(),"[","]")
    result[ErrorSeverity]= surroundWith( ErrorSeverity.String(),"[","]")
    result[FatalSeverity]= surroundWith( FatalSeverity.String(),"[","]")
    return result
}

/* Formats the input log message according to the format items.
 * Returns the resulting string. */
func formatLogMessage( msg *LogMessage, formatItems *LogFormatItems) string {
    if msg==nil {
        panic("Invalid msg argument.")
    }
    if formatItems==nil {
        panic("Invalid formatItems argument.")
    }

    var result strings.Builder
    result.Grow(defaultFormattedLogMessageCapacity)
    for _,formatItem := range *formatItems {
        switch formatItem {
            case TextFmt:
                result.WriteString( msg.text)               
            case FilenameFmt: 
                result.WriteString( msg.filename)   
            case FilenameLineFmt:           
                result.WriteString( msg.filename)   
                result.WriteString( ":")    
                result.WriteString( fmt.Sprint( msg.line))  
            case FunctNameFmt:
                result.WriteString( msg.funcName)
            case LineFmt:
                result.WriteString( fmt.Sprint( msg.line) )
            case ShortTimestampFmt:
                result.WriteString( msg.timestamp.Format(shortTimestampFormat))
            case LongTimestampFmt:
                result.WriteString( msg.timestamp.Format(longTimestampFormat))
            case SeverityFmt:
                result.WriteString( severityToText[msg.severity])
            case LineEndFmt:
                result.WriteString("\n")
        }
        if LineEndFmt!=formatItem {
            result.WriteString(" ")
        }
    }
    if result.Len()>0 {
        result.WriteString("\n")
    }

    return result.String()
}

func timestampString(t time.Time, separator string) string {
    return fmt.Sprintf("%04d%02d%02d%s%02d%02d%02d%s%03d",
                        t.Year(),
                        t.Month(),
                        t.Day(),
                        separator,
                        t.Hour(),
                        t.Minute(),
                        t.Second(),
                        separator,
                        t.Nanosecond()/1000000)
}
