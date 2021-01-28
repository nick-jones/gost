package analysis

import "github.com/nick-jones/gost/internal/pattern"

const wild = pattern.Wildcard

// directMatcher is used to match against sequences and extract pertinent information. All position values are relative to
// the start of the sequence
type directMatcher struct {
	pattern []byte // sequence of bytes to match against

	insPos int // position of the instruction that has an offset relative to the rip register

	offsetPos int // position where the offset against the rip register is set
	offsetLen int // size of the offset value in bytes

	lenPos  int // position where the string length is set
	lenSize int // size of the string length value in bytes

	// the string pointer and length are always passed around together. Because of this, we always expect to see the 2
	// values placed in adjacent memory. For direct function calls this will be the stack pointer, but in other cases
	// it may some alternative memory location. Capturing these 2 values is simply another heuristic. These are labelled
	// as args for want of a better word (this is fitting in some cases, less so in others)
	arg1Pos int
	arg2Pos int
}

var directMatchers = []directMatcher{
	{
		// first argument to a function
		pattern: []byte{
			0x48, 0x8d, 0x05, wild, wild, wild, wild, // lea rax, [rip + ????]
			0x48, 0x89, 0x04, 0x24, // mov qword ptr [rsp], rax
			0x48, 0xc7, 0x44, 0x24, 0x08, wild, wild, wild, wild, // mov qword ptr [rsp + 8], ????
		},
		insPos:    0,
		offsetPos: 3,
		offsetLen: 4,
		lenPos:    16,
		lenSize:   4,
		arg1Pos:   -1,
		arg2Pos:   15,
	},
	{
		// first argument to a function (2)
		pattern: []byte{
			0x48, 0x8d, 0x15, wild, wild, wild, wild, // lea rdx, [rip + ????]
			0x48, 0x89, 0x14, 0x24, // mov qword ptr [rsp], rdx
			0x48, 0xc7, 0x44, 0x24, 0x08, wild, wild, wild, wild, // mov  qword ptr [rsp + 8], ????
		},
		insPos:    0,
		offsetPos: 3,
		offsetLen: 4,
		lenPos:    16,
		lenSize:   4,
		arg1Pos:   -1,
		arg2Pos:   15,
	},
	{
		// first argument to a function (3)
		pattern: []byte{
			0x48, 0x8d, 0x0d, wild, wild, wild, wild, // lea rcx, [rip + ????]
			0x48, 0x89, 0x0c, 0x24, // mov qword ptr [rsp], rcx
			0x48, 0xc7, 0x44, 0x24, 0x08, wild, wild, wild, wild, // mov qword ptr [rsp + 8], ????
		},
		insPos:    0,
		offsetPos: 3,
		offsetLen: 4,
		lenPos:    16,
		lenSize:   4,
		arg1Pos:   -1,
		arg2Pos:   15,
	},
	{
		// first argument to a function (4) / concatenated
		pattern: []byte{
			0x48, 0x8d, 0x15, wild, wild, wild, wild, // lea rdx, [rip + ????]
			0x48, 0x89, 0x54, 0x24, wild, // mov qword ptr [rsp + ?], rdx
			0x48, 0xc7, 0x44, 0x24, wild, wild, wild, wild, wild, // mov qword ptr [rsp + ?], ????
		},
		insPos:    0,
		offsetPos: 3,
		offsetLen: 4,
		lenPos:    17,
		lenSize:   4,
		arg1Pos:   11,
		arg2Pos:   16,
	},
	{
		// any other argument to a function
		pattern: []byte{
			0x48, 0x8d, 0x05, wild, wild, wild, wild, // lea rax, [rip + ????]
			0x48, 0x89, 0x44, 0x24, wild, // mov qword ptr [rsp + ?], rax
			0x48, 0xc7, 0x44, 0x24, wild, wild, wild, wild, wild, // mov qword ptr [rsp + ?], ????
		},
		insPos:    0,
		offsetPos: 3,
		offsetLen: 4,
		lenPos:    17,
		lenSize:   4,
		arg1Pos:   11,
		arg2Pos:   16,
	},
	{
		// multiple string assignments
		pattern: []byte{
			0x48, 0x8d, 0x0d, wild, wild, wild, wild, // lea rcx, [rip + ????]
			0x48, 0x89, 0x0c, 0x24, // mov qword ptr [rsp], rcx
			0x48, 0xc7, 0x44, 0x24, wild, wild, wild, wild, wild, // mov qword ptr [rsp + ?], ????
			0xe8, wild, wild, wild, wild, // call ? <runtime.convTstring>
			0x48, 0x8b, 0x44, 0x24, wild, // mov rax, qword ptr [rsp + ?]
		},
		insPos:    0,
		offsetPos: 3,
		offsetLen: 4,
		lenPos:    16,
		lenSize:   4,
		arg1Pos:   15,
		arg2Pos:   29,
	},
	{
		// string comparison
		pattern: []byte{
			0x48, 0x83, 0x7c, 0x24, wild, wild, // cmp qword ptr [rsp + ?], ?
			0x74, wild, // je ?
			0xeb, wild, // jmp ?
			0x48, 0x8b, 0x44, 0x24, wild, // mov rax, qword ptr [rsp + ?]
			0x48, 0x89, 0x04, 0x24, // mov qword ptr [rsp], rax
			0x48, 0x8d, 0x05, wild, wild, wild, wild, // lea rax, [rip + ????]
		},
		insPos:    19,
		offsetPos: 22,
		offsetLen: 4,
		lenPos:    5,
		lenSize:   1,
		arg1Pos:   14,
		arg2Pos:   4,
	},
	{
		// string comparison
		pattern: []byte{
			0x48, 0x83, 0x7c, 0x24, wild, wild, // cmp qword ptr [rsp + ?], ?
			0x0f, 0x94, 0xc0, // sete al
			0x74, 0x05, // je ?
			0xe9, wild, wild, wild, wild, // jmp ?
			0x48, 0x8b, 0x44, 0x24, wild, // mov rax, qword ptr [rsp + ?]
			0x48, 0x89, 0x04, 0x24, // mov qword ptr [rsp], rax
			0x48, 0x8d, 0x05, wild, wild, wild, wild, // lea rax, [rip + ????]
		},
		insPos:    25,
		offsetPos: 28,
		offsetLen: 4,
		lenPos:    5,
		lenSize:   1,
		arg1Pos:   20,
		arg2Pos:   4,
	},
	{
		// string comparison
		pattern: []byte{
			0x48, 0x83, 0x7c, 0x24, wild, wild, // cmp qword ptr [rsp + ?], ?
			0x0f, 0x94, 0xc0, // sete al
			0x74, wild, // je ?
			0xeb, wild, // jmp ?
			0x48, 0x8b, 0x44, 0x24, wild, // mov rax, qword ptr [rsp + ?]
			0x48, 0x89, 0x04, 0x24, // mov qword ptr [rsp], rax
			0x48, 0x8d, 0x05, wild, wild, wild, wild, // lea rax, [rip + ????]
		},
		insPos:    22,
		offsetPos: 25,
		offsetLen: 4,
		lenPos:    5,
		lenSize:   1,
		arg1Pos:   17,
		arg2Pos:   4,
	},
	{
		// string into struct (1)
		pattern: []byte{
			0x48, 0xc7, 0x47, wild, wild, wild, wild, wild, // mov qword ptr [rdi + ?], ?
			0x83, 0x3d, wild, wild, wild, wild, wild, // cmp dword ptr [rip + ?????], 0
			0x0f, 0x85, wild, wild, wild, wild, // jne ????
			0x48, 0x8d, 0x0d, wild, wild, wild, wild, // lea rcx, [rip + ????]
			0x48, 0x89, 0x0f, // mov qword ptr [rdi], rcx

		},
		insPos:    21,
		offsetPos: 24,
		offsetLen: 4,
		lenPos:    4,
		lenSize:   4,
		arg1Pos:   -1,
		arg2Pos:   3,
	},
	{
		// string into struct (2)
		pattern: []byte{
			0x48, 0xc7, 0x47, wild, wild, wild, wild, wild, // mov qword ptr [rdi + ?], ?
			0x83, 0x3d, wild, wild, wild, wild, wild, // cmp dword ptr [rip + ?????], 0
			0x75, wild, // jne ?
			0x48, 0x8d, 0x05, wild, wild, wild, wild, // lea rax, [rip + ????]
			0x48, 0x89, 0x47, wild, // mov qword ptr [rdi + ?], rax
		},
		insPos:    17,
		offsetPos: 20,
		offsetLen: 4,
		lenPos:    4,
		lenSize:   4,
		arg1Pos:   27,
		arg2Pos:   3,
	},
	{
		// string into struct (3)
		pattern: []byte{
			0x48, 0xc7, 0x47, wild, wild, wild, wild, wild, // mov qword ptr [rdi + ?], ????
			0x83, 0x3d, wild, wild, wild, wild, wild, // cmp dword ptr [rip + ?????], 0
			0x0f, 0x85, wild, wild, wild, wild, // jne ????
			0x48, 0x8d, 0x0d, wild, wild, wild, wild, // lea rcx, [rip + ????]
			0x48, 0x89, 0x4f, wild, // mov qword ptr [rdi + ?], rcx
		},
		insPos:    21,
		offsetPos: 24,
		offsetLen: 4,
		lenPos:    4,
		lenSize:   4,
		arg1Pos:   31,
		arg2Pos:   3,
	},
	{
		// string into struct (4)
		pattern: []byte{
			0x48, 0xc7, 0x47, wild, wild, wild, wild, wild, // mov qword ptr [rdi + ?], ????
			0x83, 0x3d, wild, wild, wild, wild, wild, // cmp dword ptr [rip + ?????], 0
			0x0f, 0x85, wild, wild, wild, wild, // jne ????
			0x48, 0x8d, 0x05, wild, wild, wild, wild, // lea rax, [rip + ????]
			0x48, 0x89, 0x07, // mov qword ptr [rdi], rax
		},
		insPos:    21,
		offsetPos: 24,
		offsetLen: 4,
		lenPos:    4,
		lenSize:   4,
		arg1Pos:   -1,
		arg2Pos:   3,
	},
	{
		// string into struct (5)
		pattern: []byte{
			0x48, 0xc7, 0x47, wild, wild, wild, wild, wild, // mov qword ptr [rdi + ?], ????
			0x83, 0x3d, wild, wild, wild, wild, wild, // cmp dword ptr [rip + ??????], 0
			0x0f, 0x85, wild, wild, wild, wild, // jne ????
			0x48, 0x8d, 0x05, wild, wild, wild, wild, // lea rax, [rip + ????]
			0x48, 0x89, 0x47, wild, // mov qword ptr [rdi + ?], rax
		},
		insPos:    21,
		offsetPos: 24,
		offsetLen: 4,
		lenPos:    4,
		lenSize:   4,
		arg1Pos:   31,
		arg2Pos:   3,
	},
	{
		// string into struct (6)
		pattern: []byte{
			0x48, 0xc7, 0x40, wild, wild, wild, wild, wild, // mov qword ptr [rax + ?], ????
			0x48, 0x8d, 0x0d, wild, wild, wild, wild, // lea rcx, [rip + ????]
			0x48, 0x89, 0x08, // mov qword ptr [rax], rcx
		},
		insPos:    8,
		offsetPos: 11,
		offsetLen: 4,
		lenPos:    4,
		lenSize:   4,
		arg1Pos:   -1,
		arg2Pos:   3,
	},
	{
		// string into struct (7)
		pattern: []byte{
			0x48, 0x8d, 0x05, wild, wild, wild, wild, // lea rax, [rip + ????]
			0x48, 0x89, 0x84, 0x24, wild, wild, wild, wild, // mov qword ptr [rsp + ????], rax
			0x48, 0xc7, 0x84, 0x24, wild, wild, wild, wild, wild, wild, wild, wild, // mov  qword ptr [rsp + ????], ????
		},
		insPos:    0,
		offsetPos: 3,
		offsetLen: 4,
		lenPos:    23,
		lenSize:   4,
		arg1Pos:   11,
		arg2Pos:   19,
	},
	{
		// string into struct (8) - direct assignment
		pattern: []byte{
			0x48, 0xc7, 0x41, wild, wild, wild, wild, wild, // mov qword ptr [rcx + ?], ????
			0x83, 0x3d, wild, wild, wild, wild, wild, // cmp dword ptr [rip + ????], ?
			0x0f, 0x85, wild, wild, wild, wild, // jne ????
			0x48, 0x8d, 0x15, wild, wild, wild, wild, // lea rdx, [rip + ????]
			0x48, 0x89, 0x51, wild, // mov qword ptr [rcx + ?]
		},
		insPos:    21,
		offsetPos: 24,
		offsetLen: 4,
		lenPos:    4,
		lenSize:   4,
		arg1Pos:   31,
		arg2Pos:   3,
	},
	{
		// string function argument
		pattern: []byte{
			0x48, 0x8d, 0x0d, wild, wild, wild, wild, // lea rcx, [rip + ????]
			0x48, 0x89, 0x4c, 0x24, wild, // mov qword ptr [rsp + ?], rcx
			0x48, 0xc7, 0x44, 0x24, wild, wild, wild, wild, wild, // mov qword ptr [rsp + ?], ????
		},
		insPos:    0,
		offsetPos: 3,
		offsetLen: 4,
		lenPos:    17,
		lenSize:   4,
		arg1Pos:   11,
		arg2Pos:   16,
	},
	{
		// const into struct
		pattern: []byte{
			0x48, 0xc7, 0x40, wild, wild, wild, wild, wild, // mov qword ptr [rax + ?], ????
			0x48, 0x8d, 0x0d, wild, wild, wild, wild, //  lea rcx, [rip + ????]
			0x48, 0x89, 0x48, wild, // mov qword ptr [rax + ?], rcx
		},
		insPos:    8,
		offsetPos: 11,
		offsetLen: 4,
		lenPos:    4,
		lenSize:   4,
		arg1Pos:   18,
		arg2Pos:   3,
	},
	{
		// const into struct (2)
		pattern: []byte{
			0x48, 0xc7, 0x47, wild, wild, wild, wild, wild, // mov qword ptr [rdi + ?], ????
			0x48, 0x8d, 0x05, wild, wild, wild, wild, // lea rax, [rip + ????]
			0x48, 0x89, 0x47, wild, // mov qword ptr [rdi + ?], rax
		},
		insPos:    8,
		offsetPos: 11,
		offsetLen: 4,
		lenPos:    4,
		lenSize:   4,
		arg1Pos:   18,
		arg2Pos:   3,
	},
}
