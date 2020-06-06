package analysis

type indirectMatcher struct {
	pattern []byte // sequence of bytes to match against

	insPos int // position of the instruction where the value header reference is made

	typeOffsetPos int // position where the type address location is referenced
	typeOffsetLen int // size of the offset value in bytes

	valueHeaderOffsetPos int // position where the value header address location is referenced
	valueHeaderOffsetLen int // size of the offset value in bytes

	// indirect references are still typically passing values on the stack at predictable offsets, so similar
	// rules to directMatcher apply.
	arg1Pos int
	arg2Pos int
}

var indirectMatchers = []indirectMatcher{
	{
		pattern: []byte{
			0x48, 0x8d, 0x05, wild, wild, wild, wild, //  lea rax, [rip + ????]
			0x48, 0x89, 0x44, 0x24, wild, // mov qword ptr [rsp + ?], rax
			0x48, 0x8d, 0x05, wild, wild, wild, wild, // lea rax, [rip + ????]
			0x48, 0x89, 0x44, 0x24, wild, // mov qword ptr [rsp + ?], rax
		},
		insPos:               12,
		typeOffsetPos:        3,
		typeOffsetLen:        4,
		valueHeaderOffsetPos: 15,
		valueHeaderOffsetLen: 4,
		arg1Pos:              11,
		arg2Pos:              23,
	},
	{
		pattern: []byte{
			0x48, 0x8d, 0x05, wild, wild, wild, wild, // lea rax, [rip + ????]
			0x48, 0x89, 0x04, 0x24, // mov qword ptr [rsp], rax
			0x48, 0x8d, 0x05, wild, wild, wild, wild, // lea rax, [rip + ????]
			0x48, 0x89, 0x44, 0x24, wild, // mov qword ptr [rsp + ?], rax
		},
		insPos:               11,
		typeOffsetPos:        3,
		typeOffsetLen:        4,
		valueHeaderOffsetPos: 14,
		valueHeaderOffsetLen: 4,
		arg1Pos:              -1,
		arg2Pos:              22,
	},
	{
		pattern: []byte{
			0x48, 0x8d, 0x05, wild, wild, wild, wild, // lea rax, [rip + ????]
			0x48, 0x89, 0x84, 0x24, wild, wild, wild, wild, // mov qword ptr [rsp + ????]
			0x48, 0x8d, 0x05, wild, wild, wild, wild, // lea rax, [rip + ????]
			0x48, 0x89, 0x84, 0x24, wild, wild, wild, wild, // mov qword ptr [rsp + ????], rax
		},
		insPos:               15,
		typeOffsetPos:        3,
		typeOffsetLen:        4,
		valueHeaderOffsetPos: 18,
		valueHeaderOffsetLen: 4,
		arg1Pos:              11,
		arg2Pos:              26,
	},
}
