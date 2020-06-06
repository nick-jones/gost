# Go Static Strings

Static strings manifest themselves in Go compiled binaries in a number of different ways. Below are 2 deep dives into
different examples. There are of course many other examples that could be worked through.

## Example 1

Take the following Go program:

```go
package main

const x = "banana"

func main() {
	print(x)
}
```

Compile it and then decompile the main function: 

```
➜  test $ go build
➜  test $ objdump -macho -disassemble -dis-symname='_main.main' -x86-asm-syntax=intel test
test:
(__TEXT,__text) section
_main.main:
 1056e50:       65 48 8b 0c 25 30 00 00 00      mov     rcx, qword ptr gs:[48]
 1056e59:       48 3b 61 10     cmp     rsp, qword ptr [rcx + 16]
 1056e5d:       76 3b   jbe     0x1056e9a
 1056e5f:       48 83 ec 18     sub     rsp, 24
 1056e63:       48 89 6c 24 10  mov     qword ptr [rsp + 16], rbp
 1056e68:       48 8d 6c 24 10  lea     rbp, [rsp + 16]
 1056e6d:       e8 2e 36 fd ff  call    _runtime.printlock
 1056e72:       48 8d 05 d3 bf 01 00    lea     rax, [rip + 114643]
 1056e79:       48 89 04 24     mov     qword ptr [rsp], rax
 1056e7d:       48 c7 44 24 08 06 00 00 00      mov     qword ptr [rsp + 8], 6
 1056e86:       e8 55 3f fd ff  call    _runtime.printstring
 1056e8b:       e8 90 36 fd ff  call    _runtime.printunlock
 1056e90:       48 8b 6c 24 10  mov     rbp, qword ptr [rsp + 16]
 1056e95:       48 83 c4 18     add     rsp, 24
 1056e99:       c3      ret
 1056e9a:       e8 31 9d ff ff  call    _runtime.morestack_noctxt
 1056e9f:       eb af   jmp     _main.main
```

There are a few things going on here. The most relevant lines to us right now are these four:

```
 1056e72:       48 8d 05 d3 bf 01 00    lea     rax, [rip + 114643]
 1056e79:       48 89 04 24     mov     qword ptr [rsp], rax
 1056e7d:       48 c7 44 24 08 06 00 00 00      mov     qword ptr [rsp + 8], 6
 1056e86:       e8 55 3f fd ff  call    _runtime.printstring
```

Taking these line by line:

- `lea rax, [rip + 114643]` - the `lea` instruction is [Load Effective Address](https://www.aldeid.com/wiki/X86-assembly/Instructions/lea).
`rip` is the instruction pointer register. Here the `lea` instruction is being used to calculate some memory offset
_relative to the current instruction_ (or more accurately the next instruction, discussed further down). The result of
that calculation is placed into the `rax` register.
- `mov qword ptr [rsp], rax` - moves the value held into the `rax` register into the memory location pointed to by
the `rsp` register (stack pointer). This is the first argument for `runtime.printstring`.
- `mov qword ptr [rsp + 8], 6` - move `6` into the memory location pointed to by the stack pointer + 8. This is the
second argument to `runtime.printstring` and carries the length of the string.
- `call runtime.printstring` - calls the print string procedure

What do we need to understand from all of this?

1. The calling convention used by Go requires that arguments are passed on the stack. This is unlike System V x86-64
which uses a specific set of registers for arguments, only leveraging the stack once those are exhausted. It is not
covered in  these 4 lines, but Go also returns values on the stack (again different to System V x86-64).
1. The string value we're interested in loaded into memory. The offset is known at compile time, so that would suggest
the value is contained in the `__DATA` segment (most likely in the `__rodata` section - we'll check this in a moment).
1. Go passes the string length to the function - so it is more than likely that Go strings are _not_ NULL terminated.

As you can see, we have enough information in the above to work out where the string is stored in memory, and the
length of the string.

So, how do we obtain the sting value without running the executable? `objdump` is displaying the address on the left,
and conveniently we can just add the offset to this.

```
 1056e72:       48 8d 05 d3 bf 01 00    lea     rax, [rip + 114643]
 1056e79:       48 89 04 24     mov     qword ptr [rsp], rax
```

So from this we can calculate the address from where the string starts:


```
 start_address = 0x1056e79 + 114643 = 0x1072e4c
```

And since we have the length, we can also work out where it ends. In our example, the length is `6`. The start address
points to the first character, so we only need to add `5` to cover the remaining characters:

```
 start_address = 0x1056e79 + 114643 = 0x1072e4c
 end_address   = start_address + 5  = 0x1072e51
```

I mentioned before that the string is likely stored in the `__rodata` section. Now is a good moment to verify that. We
can use `objdump` to inspect this particular section:

```
➜  test $ objdump -s -j __rodata test
<snip>
 1072e40 2c206e6f 74205343 48454420 62616e61  , not SCHED bana
 1072e50 6e616566 656e6365 6f626a65 6374706f  naefenceobjectpo
<snip>
```

Let's break this down:

| address | byte | note                |
|---------|------|---------------------|
| 1072e40 | ,    |                     |
| 1072e41 |      |                     |
| 1072e42 | n    |                     |
| 1072e43 | o    |                     |
| 1072e44 | t    |                     |
| 1072e45 |      |                     |
| 1072e46 | S    |                     |
| 1072e47 | C    |                     |
| 1072e48 | H    |                     |
| 1072e49 | E    |                     |
| 1072e4a | D    |                     |
| 1072e4b |      |                     |
| 1072e4c | b    | ← our start address |
| 1072e4d | a    |                     |
| 1072e4e | n    |                     |
| 1072e4f | a    |                     |
| 1072e50 | n    |                     |
| 1072e51 | a    | ← our end address   |
| 1072e52 | e    |                     |
| 1072e53 | f    |                     |
| 1072e54 | e    |                     |
| 1072e55 | n    |                     |
| 1072e56 | c    |                     |
| 1072e57 | e    |                     |
| 1072e58 | o    |                     |
| 1072e59 | b    |                     |
| 1072e5a | j    |                     |
| 1072e5b | e    |                     |
| 1072e5c | c    |                     |
| 1072e5d | t    |                     |
| 1072e5e | p    |                     |
| 1072e5f | o    |                     |

This all lines up nicely!

So with this we have everything we need to extract string constants from the binary. Note that this only covers the
simple case where a single string argument is passed to a function; where string constants are referenced in other 
settings, the sequence of instructions will be different (but they should always at least reference the string length
and location). These won't be discussed here, as there isn't any fundamental difference in the way they location and
length can be extracted.

Now, if we experience a set of instructions that _look_ like they relate to string constants, there is always the
possibility that they don't. Thus far we only know the data resides in `__rodata` for our example, but nothing more.
`__rodata` can be used for a number of things, so there is no guarantee we are dealing with a string constant. So, we
might need another heuristic. Fortunately for us, Go adds a handy symbol that indicates which chunk of data relates
to constant strings - this is named `go.string.*`. We can see the address where this starts by using the `-t` flag for
`objdump`:

```
➜  test $ objdump -t test | rg 'go\.string\.\*'
0000000001072c38 l     O __TEXT,__rodata        _go.string.*
```

So this tells us the `go.string.*` data starts at `1072c38`. Let's have a look at `__rodata` again:

```
➜  test $ objdump -s -j __rodata test
<snip>
 1072c30 81060000 00000000 2028292b 2c2d2e2f  ........ ()+,-./
 1072c40 3a3c3d3f 5b0a095d 202b2040 2050205b  :<=?[..] + @ P [
 1072c50 2920290a 2c202d3e 3a203e20 220a0a20  ) )., ->: > "..
<snip>
```

The address indicated is towards the end of the first line. So that looks about right - a bunch of ASCII characters
starting at that particular address. So, armed with this, we have an additional heuristic; if an instruction references
addresses within that block, it's highly likely a string. If it is outside that block, we can ignore it.

Armed with all of this, we can do a pretty good job of extracting strings. Unfortunately not all cases appear like
this..

## Example 2

Another program that is identical to the 1st example in terms of what it does:

```go
package main

import "fmt"

const x = "banana"

func main() {
	fmt.Print(x)
}
```

Build and decompile:

```
➜  test $ go build                             
➜  test $ objdump -macho -disassemble -dis-symname='_main.main' -x86-asm-syntax=intel test
test:
(__TEXT,__text) section
_main.main:
 109cfa0:       65 48 8b 0c 25 30 00 00 00      mov     rcx, qword ptr gs:[48]
 109cfa9:       48 3b 61 10     cmp     rsp, qword ptr [rcx + 16]
 109cfad:       76 70   jbe     0x109d01f
 109cfaf:       48 83 ec 58     sub     rsp, 88
 109cfb3:       48 89 6c 24 50  mov     qword ptr [rsp + 80], rbp
 109cfb8:       48 8d 6c 24 50  lea     rbp, [rsp + 80]
 109cfbd:       0f 57 c0        xorps   xmm0, xmm0
 109cfc0:       0f 11 44 24 40  movups  xmmword ptr [rsp + 64], xmm0
 109cfc5:       48 8d 05 74 e2 00 00    lea     rax, [rip + 57972]
 109cfcc:       48 89 44 24 40  mov     qword ptr [rsp + 64], rax
 109cfd1:       48 8d 05 28 b8 04 00    lea     rax, [rip + 309288]
 109cfd8:       48 89 44 24 48  mov     qword ptr [rsp + 72], rax
 109cfdd:       48 8b 05 94 e0 0d 00    mov     rax, qword ptr [rip + _os.Stdout]
 109cfe4:       48 8d 0d 95 d0 04 00    lea     rcx, [rip + "_go.itab.*os.File,io.Writer"]
 109cfeb:       48 89 0c 24     mov     qword ptr [rsp], rcx
 109cfef:       48 89 44 24 08  mov     qword ptr [rsp + 8], rax
 109cff4:       48 8d 44 24 40  lea     rax, [rsp + 64]
 109cff9:       48 89 44 24 10  mov     qword ptr [rsp + 16], rax
 109cffe:       48 c7 44 24 18 01 00 00 00      mov     qword ptr [rsp + 24], 1
 109d007:       48 c7 44 24 20 01 00 00 00      mov     qword ptr [rsp + 32], 1
 109d010:       e8 8b 99 ff ff  call    _fmt.Fprint
 109d015:       48 8b 6c 24 50  mov     rbp, qword ptr [rsp + 80]
 109d01a:       48 83 c4 58     add     rsp, 88
 109d01e:       c3      ret
 109d01f:       e8 8c c4 fb ff  call    _runtime.morestack_noctxt
 109d024:       e9 77 ff ff ff  jmp     _main.main
```

This is.. quite different. First let's identify the relevant section:

```
 109cfc5:       48 8d 05 74 e2 00 00    lea     rax, [rip + 57972]
 109cfcc:       48 89 44 24 40  mov     qword ptr [rsp + 64], rax
 109cfd1:       48 8d 05 28 b8 04 00    lea     rax, [rip + 309288]
 109cfd8:       48 89 44 24 48  mov     qword ptr [rsp + 72], rax
```

You'll notice these lines don't contain anything indicating a string length. It's not entirely clear, but the
signature of `fmt.Print` may give us a clue:

```go
func Print(a ...interface{}) (n int, err error)
```

So it receives empty interfaces as arguments. Given the way interfaces work, it's entirely possible what we're seeing
here is something that indicates type along with a pointer to the underlying value. Let's check that first reference
and run with the theory that it is carrying some type indication:

```
 109cfc5:       48 8d 05 74 e2 00 00    lea     rax, [rip + 57972]
 109cfcc:       48 89 44 24 40  mov     qword ptr [rsp + 64], rax
```

```
 address = 0x109cfcc + 57972 = 0x10ab240
```

Looking this up we find:

```
<snip>
 10ab240 10000000 00000000 08000000 00000000  ................
 10ab250 b45cffe0 07080818 20460d01 00000000  .\...... F......
 10ab260 a0630e01 00000000 64170000 e0b20000  .c......d.......
<snip>
```

A bit of sleuthing shows up https://golang.org/src/runtime/typekind.go as a potential source of type identifiers. Here
`kindString` is `24` (`0x18`). In that block of data, there is only one `0x18`, at address `0x10ab257` (+23 from the
original address). Perhaps that is holding the type kind, let's test our theory with a different type:

```go
package main

import (
	"fmt"
)

const x uint16 = 60000

func main() {
	fmt.Print(x)
}
```

```
➜  test $ go build
➜  test $ objdump -macho -disassemble -dis-symname='_main.main' -x86-asm-syntax=intel test
test:
(__TEXT,__text) section
_main.main:
 109cfa0:       65 48 8b 0c 25 30 00 00 00      mov     rcx, qword ptr gs:[48]
 109cfa9:       48 3b 61 10     cmp     rsp, qword ptr [rcx + 16]
 109cfad:       76 70   jbe     0x109d01f
 109cfaf:       48 83 ec 58     sub     rsp, 88
 109cfb3:       48 89 6c 24 50  mov     qword ptr [rsp + 80], rbp
 109cfb8:       48 8d 6c 24 50  lea     rbp, [rsp + 80]
 109cfbd:       0f 57 c0        xorps   xmm0, xmm0
 109cfc0:       0f 11 44 24 40  movups  xmmword ptr [rsp + 64], xmm0
 109cfc5:       48 8d 05 f4 e2 00 00    lea     rax, [rip + 58100]
 109cfcc:       48 89 44 24 40  mov     qword ptr [rsp + 64], rax
 109cfd1:       48 8d 05 ce b2 04 00    lea     rax, [rip + 307918]
 109cfd8:       48 89 44 24 48  mov     qword ptr [rsp + 72], rax
 109cfdd:       48 8b 05 94 e0 0d 00    mov     rax, qword ptr [rip + _os.Stdout]
 109cfe4:       48 8d 0d 75 d0 04 00    lea     rcx, [rip + "_go.itab.*os.File,io.Writer"]
 109cfeb:       48 89 0c 24     mov     qword ptr [rsp], rcx
 109cfef:       48 89 44 24 08  mov     qword ptr [rsp + 8], rax
 109cff4:       48 8d 44 24 40  lea     rax, [rsp + 64]
 109cff9:       48 89 44 24 10  mov     qword ptr [rsp + 16], rax
 109cffe:       48 c7 44 24 18 01 00 00 00      mov     qword ptr [rsp + 24], 1
 109d007:       48 c7 44 24 20 01 00 00 00      mov     qword ptr [rsp + 32], 1
 109d010:       e8 8b 99 ff ff  call    _fmt.Fprint
 109d015:       48 8b 6c 24 50  mov     rbp, qword ptr [rsp + 80]
 109d01a:       48 83 c4 58     add     rsp, 88
 109d01e:       c3      ret
 109d01f:       e8 8c c4 fb ff  call    _runtime.morestack_noctxt
 109d024:       e9 77 ff ff ff  jmp     _main.main
```

```
 address = 0x109cfcc + 58100 = 0x10ab2c0
```

```
 10ab2c0 02000000 00000000 00000000 00000000  ................
 10ab2d0 a00ef2ef 0f020209 58440d01 00000000  ........XD......
 10ab2e0 98630e01 00000000 6e170000 e0b50000  .c......n.......
```

So let's add 23 to our address:

```
0x10ab2c0 + 23 = 0x10ab2d7
```

This address holds a value of `0x09`. `kindUint16` is also `0x09`!

A bit more sleuthing suggests this block of memory may well be [`runtime._type`](https://golang.org/src/runtime/type.go#L31)
or some variation to it. Counting the bytes for each field:

```go
type _type struct {
	size       uintptr // +8 bytes
	ptrdata    uintptr // +8 bytes
	hash       uint32 // +4 bytes
	tflag      tflag // +1 byte
	align      uint8 // +1 byte
	fieldAlign uint8 // +1 byte
	kind       uint8
	// <snip>
}
```

23 bytes to reach `kind`.

So we have a likely way to obtain the type. So what about the second reference? Could this be the string value? Let's
check:

```
 109cfd1:       48 8d 05 28 b8 04 00    lea     rax, [rip + 309288]
 109cfd8:       48 89 44 24 48  mov     qword ptr [rsp + 72], rax
```

```
 address = 0x109cfd8 + 309288 = 0x10e8800
```

```
 10e8800 5ace0c01 00000000 06000000 00000000  Z...............
```

No `banana`! Well actually that makes sense. As discussed in Example 1, Go always wants to deal with strings by carrying
a pointer and length. So perhaps that's what we have here? The most obvious thing to look for is the length of our
string, which is 6. We can see that in the latter 8 bytes. This is Mach-O and little endian, so we need to do a bit of
re-arranging to get those 8 bytes as big endian:

```
little endian: 0600000000000000
big endiate:   0000000000000006
```

OK, so a value of 6! That matches nicely.

So the other 8 bytes may well be a pointer, so let's check that out. Switching to big endian again:

```
little endian = 5ace0c0100000000
big endian    = 00000000010cce5a
```

So a potential address of `0x10cce5a`

```
 10cce50 6e63686f 5b5d6279 74656261 6e616e61  ncho[]bytebanana
 10cce60 6368616e 3c2d6566 656e6365 6572726e  chan<-efenceerrn
```

`banana` located! With all of this we have another means to locate strings. Unfortunately the block of data holding
string pointer and length does not have any obvious symbols assigned. (I believe this area is referred
to as `statictmp` or `stmp` - some earlier versions of Go did leave symbols here, newer versions do not.
[There may be some hope for them making a return](https://github.com/golang/go/issues/39053).)
