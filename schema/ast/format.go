package ast

import (
	"fmt"
	"io"
)

type bailout error

// Format will format the schema into the canonical human-readable format
// that matches s. There are additional features supported in the human readable
// format that will be lost in the JSON format, such as comments and additional formatting.
func Format(n Node, out io.Writer) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if bail, ok := r.(bailout); ok {
				err = bail
			}
		}
	}()
	p := &formatter{
		w:        out,
		lastchar: '\n',
		tab:      "    ", // 4 spaces
	}
	p.print(n)
	return nil
}

type formatter struct {
	indent int // 1 = 1 tab over
	w      io.Writer

	lastchar byte
	tab      string
}

func (p *formatter) printInd(s string) {
	if p.lastchar != '\n' {
		panic("lastchar must be newline when calling printInd")
	}
	for i := 0; i < p.indent; i++ {
		p.write(p.tab)
	}
	if len(s) > 0 {
		p.write(s)
		p.lastchar = s[len(s)-1]
	}
}

func (p *formatter) printIndf(format string, args ...any) {
	p.printInd(fmt.Sprintf(format, args...))
}

func (p *formatter) write(s string) {
	_, err := io.WriteString(p.w, s)
	if err != nil {
		panic(bailout(err))
	}
	p.lastchar = s[len(s)-1]
}

func (p *formatter) writef(format string, args ...any) {
	buf := fmt.Sprintf(format, args...)
	p.write(buf)
}

func (p *formatter) print(n Node) {
	switch n := n.(type) {
	case *Schema:
		for _, d := range n.Decls {
			p.print(d)
		}
		p.print(n.FooterComments)
	case *Namespace:
		p.print(n.Before)
		p.printInd("namespace ")
		p.print(n.Name)
		p.write(" {")
		if n.Inline != nil {
			p.print(n.Inline)
		}
		p.write("\n")
		for _, d := range n.Decls {
			p.indent += 1
			p.print(d)
			p.indent -= 1
		}
		p.write("}")
		if n.Footer != nil {
			p.print(n.Footer)
		}
		p.write("\n")
	case *CommonTypeDecl:
		p.print(n.Before)
		p.printIndf("type %s = ", n.Name.Value)
		p.print(n.Value)
		p.write(";")
		if n.Footer != nil {
			p.print(n.Footer)
		}
		p.write("\n")
	case *RecordType:
		p.write("{")
		if n.Inner != nil {
			p.print(n.Inner)
		}
		p.write("\n")
		for _, a := range n.Attributes {
			p.indent += 1
			p.print(a)
			p.indent -= 1
		}
		if len(n.Remaining) > 0 {
			p.indent += 1
			p.print(n.Remaining)
			p.indent -= 1
		}
		p.printInd("}")
	case *SetType:
		p.write("Set<")
		p.print(n.Element)
		p.write(">")
	case *Path:
		for i, part := range n.Parts {
			if i > 0 {
				p.write("::")
			}
			p.print(part)
		}
	case *Attribute:
		p.print(n.Before)
		p.printInd("") // print indent
		p.print(n.Key)
		if !n.IsRequired {
			p.write("?")
		}
		p.write(": ")
		p.print(n.Type)
		p.write(",")
		if n.Inline != nil {
			p.print(n.Inline)
		}
		p.write("\n")
	case *Ident:
		p.write(n.Value)
	case *String:
		p.write(n.QuotedVal)
	case *Entity:
		p.print(n.Before)
		p.printInd("entity ")
		for i, name := range n.Names {
			if i > 0 {
				p.write(", ")
			}
			p.print(name)
		}
		if n.In != nil {
			p.write(" in ")
			printBracketList(p, n.In)
		}
		if n.Shape != nil {
			if n.EqTok.Line > 0 {
				p.write(" = ")
			} else {
				p.write(" ")
			}
			p.print(n.Shape)
		}
		if n.Tags != nil {
			p.write(" tags ")
			p.print(n.Tags)
		}
		p.write(";")
		if n.Footer != nil {
			p.print(n.Footer)
		}
		p.write("\n")
	case *Action:
		p.print(n.Before)
		p.printInd("action ")
		for i, name := range n.Names {
			if i > 0 {
				p.write(", ")
			}
			p.print(name)
		}
		if len(n.In) > 0 {
			p.write(" in ")
			printBracketList(p, n.In)
		}
		if n.AppliesTo != nil {
			p.write(" appliesTo {")
			if n.AppliesTo.Inline != nil {
				p.print(n.AppliesTo.Inline)
			}
			p.write("\n")
			p.indent += 1
			p.print(n.AppliesTo)
			p.indent -= 1
			p.printInd("}")
		}
		p.write(";")
		if n.Footer != nil {
			p.print(n.Footer)
		}
		p.write("\n")
	case *AppliesTo:
		if len(n.Principal) > 0 {
			p.print(n.PrincipalComments.Before)
			p.printInd("principal: ")
			printBracketList(p, n.Principal)
			p.write(",")
			if n.PrincipalComments.Inline != nil {
				p.print(n.PrincipalComments.Inline)
			}
			p.write("\n")
		}
		if len(n.Resource) > 0 {
			p.print(n.ResourceComments.Before)
			p.printInd("resource: ")
			printBracketList(p, n.Resource)
			p.write(",")
			if n.ResourceComments.Inline != nil {
				p.print(n.ResourceComments.Inline)
			}
			p.write("\n")
		}
		if n.Context != nil {
			p.print(n.ContextComments.Before)
			p.printInd("context: ")
			p.print(n.Context)
			p.write(",")
			if n.ContextComments.Inline != nil {
				p.print(n.ContextComments.Inline)
			}
			p.write("\n")
		}
		p.print(n.Remaining)
	case *Ref:
		for i, part := range n.Namespace {
			if i > 0 {
				p.write("::")
			}
			p.print(part)
		}
		if len(n.Namespace) > 0 {
			p.write("::")
		}
		p.print(n.Name)
	case CommentBlock:
		if len(n) == 0 {
			return
		}
		for _, c := range n {
			// Print each comment line on a separate line indented
			p.printInd("")
			p.print(c)
			p.write("\n")
		}
	case *Comment:
		if p.lastchar != ' ' && p.lastchar != '\t' && p.lastchar != '\x00' && p.lastchar != '\n' {
			p.write(" ")
		}
		p.writef("// %s", n.Trim())
	default:
		panic(fmt.Sprintf("unhandled node type %T", n))
	}
}

func printBracketList[T Node](p *formatter, list []T) {
	if len(list) == 0 {
		panic("list must not be empty")
	}
	if len(list) > 1 {
		p.write("[")
	}
	for i, item := range list {
		if i > 0 {
			p.write(", ")
		}
		p.print(item)
	}
	if len(list) > 1 {
		p.write("]")
	}
}
