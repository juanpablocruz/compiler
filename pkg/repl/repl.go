package repl

import (
	"bufio"
	"fmt"
	"io"

	"github.com/juanpablocruz/monkey/pkg/compiler"
	"github.com/juanpablocruz/monkey/pkg/evaluator"
	"github.com/juanpablocruz/monkey/pkg/lexer"
	"github.com/juanpablocruz/monkey/pkg/object"
	"github.com/juanpablocruz/monkey/pkg/parser"
	"github.com/juanpablocruz/monkey/pkg/vm"
)

const PROMPT = ">> "

func StartCompiled(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	constants := []object.Object{}
	globals := make([]object.Object, vm.GlobalsSize)
	symbolTable := compiler.NewSymbolTable()

	for i, v := range object.Builtins {
		symbolTable.DefineBuiltin(i, v.Name)
	}

	for {
		fmt.Fprint(out, PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()

		if line == "quit" {
			return
		}
		l := lexer.New(line)
		p := parser.New(l)

		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			printParserErrors(out, p.Errors())
			continue
		}

		comp := compiler.NewWithState(symbolTable, constants)
		err := comp.Compile(program)
		if err != nil {
			fmt.Fprintf(out, "Woops! Compilation failed:\n %s\n", err)
			continue
		}

		machine := vm.NewWithGlobalsStore(comp.Bytecode(), globals)
		err = machine.Run()
		if err != nil {
			fmt.Fprintf(out, "Woops! Executing bytecode failed:\n %s\n", err)
			continue
		}

		stackTop := machine.LastPoppedStackElem()
		io.WriteString(out, stackTop.Inspect())
		io.WriteString(out, "\n")

	}
}

func StartInterpreted(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	env := object.NewEnvironment()
	macroEnv := object.NewEnvironment()

	for {
		fmt.Fprint(out, PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()

		if line == "quit" {
			return
		}
		l := lexer.New(line)
		p := parser.New(l)

		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			printParserErrors(out, p.Errors())
			continue
		}

		evaluator.DefineMacros(program, macroEnv)
		expanded := evaluator.ExpandMacros(program, macroEnv)

		evaluated := evaluator.Eval(expanded, env)

		if evaluated != nil {
			io.WriteString(out, evaluated.Inspect())
			io.WriteString(out, "\n")
		}
	}
}

func printParserErrors(out io.Writer, errors []string) {
	io.WriteString(out, "Woops! We ran into some monkey business here!\n")
	io.WriteString(out, " parser errors:\n")
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
}
