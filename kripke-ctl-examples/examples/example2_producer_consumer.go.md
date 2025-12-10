===============================================================================
 Producer-Consumer with Bounded Buffer - Complete Example
===============================================================================

SYSTEM DESCRIPTION
------------------------------------------------------------------------------

Producer-Consumer with bounded buffer (capacity = 2):
- Producer creates items and sends them to a buffer
- Consumer receives items from the buffer
- When buffer is full (2 items), producer must wait
- When buffer is empty, consumer must wait

STATE SPACE
------------------------------------------------------------------------------
States (3 total):
  B1: [P-ready, C-ready]
  B2: [full, C-ready]
  B0: [empty, P-ready]

CTL PROPERTIES
------------------------------------------------------------------------------
✓ PASS P1: Buffer never overflows (should pass)
✓ PASS P2: Always possible to eventually produce
✓ PASS P3: Always possible to eventually consume
✓ PASS P4: System never deadlocks
✓ PASS P5: Buffer can become full
✓ PASS P6: Buffer can become empty

MERMAID DIAGRAM
------------------------------------------------------------------------------
stateDiagram-v2
    [*] --> B0
    
    B0 --> B1: produce
    B1 --> B2: produce
    B1 --> B0: consume
    B2 --> B1: consume
    
    B0: Empty (P:ready C:blocked)
    B1: Half (P:ready C:ready)
    B2: Full (P:blocked C:ready)


===============================================================================
