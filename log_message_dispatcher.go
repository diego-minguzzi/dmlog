package dmlog

import "fmt"

type replyType struct {
    ok bool
}

// Message Sink - request message.
type reqMessageSinkType struct {
    messageSink *LogMessageSink
}

// Message Sink - reply message.
type replyMessageSinkType struct {
    replyType
    sinkId   MessageSinkId  
}

// Message sink threshold - request message.
type reqMessageSinkThresholdType struct {
    sinkId   MessageSinkId
    threshold LogSeverity
}

// Message sink threshold - reply message.
type replyMessageSinkThresholdType struct {
    replyType
}

// Clear sinks : request message.
type reqClearSinksType struct {}

// Clear sinks : reply message.
type replyClearSinksType struct {
    replyType
}

// Set sink format - request message.
type reqSetSinkFormatType struct {
    sinkId      MessageSinkId
    messageType MessageType
    formatItems LogFormatItems
}

// Set sink format - reply message.
type replySetSinkFormatType struct {
    replyType
}

/* Issues a request that sets the format of for a message type of a given sink.
 * The sink is identified by the sinkId, that must be previously added.
 * formatItems is a sequence of LogFormatItem elements.
 */
func reqSetSinkFormat(sinkId MessageSinkId, 
                      messageType MessageType, 
                      formatItems ...LogFormatItem) bool {
    context.chRequest <- reqSetSinkFormatType { 
                            sinkId: sinkId,
                            messageType: messageType,
                            formatItems: formatItems,
                        }
    switch reply := (<- context.chReply).(type) {
        case replySetSinkFormatType: {
            return reply.ok
        }       
        default:
            panic(fmt.Sprintf("unexpected reply type %T",reply))
    }
}

//--------------------------------------------------------------------------------------------------
// Issues a request to remove all sinks.  It blocks waiting for the result. 
func reqClearSinks() bool {
    context.chRequest <- reqClearSinksType{ }
    switch reply := (<- context.chReply).(type) {
        case replyClearSinksType: {
            return reply.ok
        }       
        default:
            panic(fmt.Sprintf("unexpected reply type %T",reply))
    }
}

//--------------------------------------------------------------------------------------------------
// Issues a request to add a message sink.  It blocks waiting for the result. 
func reqMessageSink( messageSink *LogMessageSink) (MessageSinkId, error) {
    if messageSink == nil {
        return MessageSinkId(0), fmt.Errorf("reqMessageSink(): invalid argument")
    }
    context.chRequest <- reqMessageSinkType{ messageSink:messageSink,}
    switch reply := (<- context.chReply).(type) {
        case replyMessageSinkType: {
            if ! reply.ok {
              return MessageSinkId(0),fmt.Errorf("failed")
            }
            return reply.sinkId, nil
        }       
        default:
            panic(fmt.Sprintf("unexpected reply type %T",reply))
    }
} 

//--------------------------------------------------------------------------------------------------
// Issues a request to set a sink threshold.  It blocks waiting for the result. 
func reqMessageSinkThreshold( sinkId MessageSinkId, threshold LogSeverity) bool {
    context.chRequest <- reqMessageSinkThresholdType{ sinkId:sinkId, threshold:threshold,}
    switch reply := (<- context.chReply).(type) {
        case replyMessageSinkThresholdType: {
            return reply.ok
        }       
        default:
            panic(fmt.Sprintf("unexpected reply type %T",reply))
    }
}

type ctxMessageDispatcher struct {
    sinks []*LogMessageSink
}

//--------------------------------------------------------------------------------------------------
func messageDispatcher() {
    var ctx = ctxMessageDispatcher{ sinks: make([]*LogMessageSink, 0, defaultSinksCapacity), }
        
    for isTerminate:=false; !isTerminate; {
        select {
            case newMessage := <- context.chLogMessages: {
                for _, sink := range ctx.sinks {
                    (*sink).OnLogMessage( &newMessage)
                }
            }

            case newRequest := <- context.chRequest: {
                context.chReply <- handleRequest( newRequest, &ctx)
            }

            case <- context.chReqTerminate: {
                stillHasMessages := true
                for stillHasMessages {   
                    // Flushes all pending messages.
                    select {
                        case newMessage := <- context.chLogMessages: {
                            for _, sink := range ctx.sinks {
                                (*sink).OnLogMessage( &newMessage)
                            }
                        }
                        default:
                            stillHasMessages= false
                    }
                }
                for _, sink := range ctx.sinks {
                    (*sink).terminate()
                }
                close(context.chReplyTerminate)
                isTerminate = true                
            }    
        }
    }
}

//--------------------------------------------------------------------------------------------------
func handleRequest( request interface{}, ctx *ctxMessageDispatcher) interface{} {
    switch request := request.(type) {
        case reqMessageSinkType: {
            ctx.sinks= append(ctx.sinks, request.messageSink )
            newSinkId := MessageSinkId( len(ctx.sinks)-1 )                        
            return replyMessageSinkType{ replyType{true}, newSinkId}
        }
        case reqMessageSinkThresholdType: {
            for indx, sink := range ctx.sinks {
                if MessageSinkId(indx)== request.sinkId {
                    (*sink).SetSeverity( request.threshold)
                    return replyMessageSinkThresholdType{ replyType{true} }
                }                   
            }
            return replyMessageSinkThresholdType{ replyType{false} }
        }
        case reqClearSinksType: {
            for _, sink := range ctx.sinks {
                    (*sink).terminate()
            }
            ctx.sinks= make([]*LogMessageSink, 0, defaultSinksCapacity)

            return replyClearSinksType{ replyType{true}, }
        }
        case reqSetSinkFormatType: {
            for indx, sink := range ctx.sinks {
                if MessageSinkId(indx)== request.sinkId {
                    isOk := (*sink).setSinkFormat( request.messageType, request.formatItems)
                    return replySetSinkFormatType{ replyType{isOk}, }
                }                   
            }
            return replySetSinkFormatType{ replyType{false}, }
        }  
        default : {
            return replyType{false}
        }
    }
}
