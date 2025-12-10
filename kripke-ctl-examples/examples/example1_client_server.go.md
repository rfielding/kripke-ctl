===============================================================================
 Client-Server System with Timeout - Complete Example
===============================================================================

SYSTEM DESCRIPTION
------------------------------------------------------------------------------

A client-server system with timeout:
- Client sends requests to the server
- Server processes requests and sends responses
- If no response arrives within 3 time units, the client times out
- After timeout or response, the client can send a new request

STATE SPACE
------------------------------------------------------------------------------
States (6 total):
  waiting_processing_1
  waiting_processing_2
  idle_idle_response
  idle_idle_timeout
  idle_idle_0
  waiting_processing_0

CTL PROPERTIES
------------------------------------------------------------------------------
✓ PASS P1: Client is always either idle or waiting
✓ PASS P2: It's possible to get a response
✓ PASS P3: It's possible to timeout
✓ PASS P4: Never both response and timeout

MERMAID DIAGRAM
------------------------------------------------------------------------------
stateDiagram-v2
    [*] --> idle_idle_0
    
    waiting_processing_0 --> waiting_processing_1
    waiting_processing_1 --> waiting_processing_2
    waiting_processing_1 --> idle_idle_response
    waiting_processing_2 --> idle_idle_response
    waiting_processing_2 --> idle_idle_timeout
    idle_idle_response --> idle_idle_0
    idle_idle_timeout --> idle_idle_0
    idle_idle_0 --> waiting_processing_0
    
    idle_idle_0: Idle / Ready
    waiting_processing_0: Sent / Timer=0
    waiting_processing_1: Wait / Timer=1
    waiting_processing_2: Wait / Timer=2
    idle_idle_response: Response Received
    idle_idle_timeout: Timeout!


===============================================================================
