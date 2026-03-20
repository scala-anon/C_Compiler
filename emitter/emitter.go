package emitter  

import "os"

type Emitter struct {
	FullPath   string
	Data       string    // .data section
	Bss        string    // .bss section
	Code       string    // .text section
	StackOffset int      // current stack offset
	Variables  map[string]int  // variable name → stack offset
	LabelCount int       // for generating unique labels
}

func NewEmitter(fullPath string) *Emitter {
	return &Emitter{
		FullPath:    fullPath,
		StackOffset: 0,
		Variables:   make(map[string]int),
		LabelCount:  0,
	}
}

/*Data — strings go here. When you write "hello" in your C code, the emitter adds str_0: db "hello", 0 to this section.

Bss — reserved space with no initial value. We might not use this much since local variables live on the stack.

Code — the actual instructions. This is where mov, add, cmp, jmp all go.

StackOffset — starts at 0. Every time you declare a variable, it decreases by 4 (one int = 4 bytes). So first variable is at -4, second at -8, etc.

Variables — maps names to their stack position. When you see int x = 5;, you store "x" → -4. Later when you see x + 1, you look up "x" and know to emit mov eax, [rbp-4].

LabelCount — every if/while/for needs unique labels. Each time you need one, increment this and use it: .L0, .L1, .L2...

Then the helper methods:
*/

// Emit to .text section
func (e *Emitter) Emit(code string) {
	e.Code += code
}

func (e *Emitter) EmitLine(code string) {
	e.Code += "    " + code + "\n"
}

// Emit a label (no indent)
func (e *Emitter) EmitLabel(label string) {
	e.Code += label + ":\n"
}

// Emit to .data section
func (e *Emitter) DataLine(code string) {
	e.Data += "    " + code + "\n"
}

// Get next unique label number
func (e *Emitter) NextLabel() int {
	e.LabelCount++
	return e.LabelCount
}

// Declare a variable on the stack, return its offset
func (e *Emitter) DeclareVariable(name string) int {
	e.StackOffset -= 4
	e.Variables[name] = e.StackOffset
	return e.StackOffset
}

// Look up a variable's stack offset
func (e *Emitter) GetVariable(name string) int {
	return e.Variables[name]
}

// Write the final .asm file
func (e *Emitter) WriteFile() {
	output := "section .note.GNU-stack noalloc noexec nowrite progbits\n\n"
	output += "section .data\n"
	output += e.Data
	output += "\nsection .bss\n"
	output += e.Bss
	output += "\nsection .text\n"
	output += e.Code
	os.WriteFile(e.FullPath, []byte(output), 0644)
}
