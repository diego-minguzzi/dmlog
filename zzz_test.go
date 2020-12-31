package dmlog

import "testing"
import "time"

//--------------------------------------------------------------------------------------------------
func TestTerminated( t *testing.T) {    
    if IsTerminated() {
        t.Error(t.Name(),`IsTerminated(): got true, expected false`)            
    }
    
    time.Sleep( 500*time.Millisecond)
    Terminate()
    if ! IsTerminated() {
        t.Error(t.Name(),`IsTerminated(): got false, expected true`)            
    }    
}
