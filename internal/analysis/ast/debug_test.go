package ast

import (
	"context"
	"fmt"
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/stretchr/testify/require"
)

func TestDebugInterfaceAST(t *testing.T) {
	parser, err := NewParser()
	require.NoError(t, err)
	defer parser.Close()

	code := `interface Shape extends Drawable {
    area(): number;
}`

	// Parse directly to examine AST structure
	tree, err := parser.tsParser.ParseCtx(context.Background(), nil, []byte(code))
	require.NoError(t, err)
	defer tree.Close()

	printAST(tree.RootNode(), []byte(code), 0)
}

func TestDebugTypeScriptClass(t *testing.T) {
	parser, err := NewParser()
	require.NoError(t, err)
	defer parser.Close()

	code := `class UserService {
    private apiUrl: string;
    
    constructor(apiUrl: string) {
        this.apiUrl = apiUrl;
    }
}`

	// Parse directly to examine AST structure
	tree, err := parser.tsParser.ParseCtx(context.Background(), nil, []byte(code))
	require.NoError(t, err)
	defer tree.Close()

	printAST(tree.RootNode(), []byte(code), 0)
}

func printAST(node *sitter.Node, content []byte, depth int) {
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}

	nodeText := string(content[node.StartByte():node.EndByte()])
	if len(nodeText) > 50 {
		nodeText = nodeText[:47] + "..."
	}

	fmt.Printf("%s%s: '%s'\n", indent, node.Type(), nodeText)

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		printAST(child, content, depth+1)
	}
}
