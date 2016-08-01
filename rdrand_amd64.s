// func rdrand() uint64
TEXT ·rdrand(SB), $0-16
	BYTE $0x48; BYTE $0x0F; BYTE $0xC7; BYTE $0xF0 // RDRANDQ AX
	MOVQ AX, ret+0(FP)
	RET
