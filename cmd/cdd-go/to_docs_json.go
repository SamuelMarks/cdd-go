package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/samuel/cdd-go/src/openapi"
)

type DocsJSONOutput struct {
        Endpoints map[string]map[string]string `json:"endpoints"`
}

func runToDocsJSON(args []string) error {
        fs := flag.NewFlagSet("to_docs_json", flag.ContinueOnError)
        fs.SetOutput(os.Stderr)

        var in string
        var out string
        var noImports bool
        var noWrapping bool

        fs.StringVar(&in, "i", envOrDefault("CDD_GO_INPUT", ""), "Input file path")
        fs.StringVar(&in, "input", envOrDefault("CDD_GO_INPUT", ""), "Input file path")
        fs.StringVar(&out, "o", envOrDefault("CDD_GO_OUTPUT", ""), "Output file path")
        fs.BoolVar(&noImports, "no-imports", envOrDefaultBool("CDD_GO_NO_IMPORTS", false), "Omit imports")
        fs.BoolVar(&noWrapping, "no-wrapping", envOrDefaultBool("CDD_GO_NO_WRAPPING", false), "Omit wrapping")

        if err := fs.Parse(args); err != nil {
                return err
        }

        if in == "" {
                return fmt.Errorf("input file is required")
        }

        f, err := os.Open(in)
        if err != nil {
                return err
        }
        defer f.Close()

        oa, err := openapi.Parse(f)
        if err != nil {
                return err
        }

        endpoints := make(map[string]map[string]string)
        
        for path, item := range oa.Paths {
                methods := map[string]*openapi.Operation{
                        "get":     item.Get,
                        "put":     item.Put,
                        "post":    item.Post,
                        "delete":  item.Delete,
                        "options": item.Options,
                        "head":    item.Head,
                        "patch":   item.Patch,
                        "trace":   item.Trace,
                }

                pathMap := make(map[string]string)

                for method, op := range methods {
                        if op == nil {
                                continue
                        }

                        opID := op.OperationID
                        if opID == "" {
                            opID = "request"
                        }

                        finalCode := ""
                        if !noImports {
                                finalCode += "import (\n\t\"context\"\n\t\"fmt\"\n\t\"os\"\n\t\"github.com/your/client\"\n)\n\n"
                        }
                        
                        if !noWrapping {
                                finalCode += "func main() {\n"
                        }

                        finalCode += fmt.Sprintf(`	client := client.NewAPIClient(client.NewConfiguration())
	resp, r, err := client.DefaultApi.%s(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %%v
", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stdout, "Response: %%v
", resp)`, opID)

                        if !noWrapping {
                                finalCode += "\n}"
                        }
                        
                        pathMap[method] = finalCode
                }
                
                if len(pathMap) > 0 {
                    endpoints[path] = pathMap
                }
        }

        result := DocsJSONOutput{
                Endpoints: endpoints,
        }

        outTarget := os.Stdout
        if out != "" {
                fileTarget, err := os.Create(out)
                if err != nil {
                        return err
                }
                defer fileTarget.Close()
                outTarget = fileTarget
        }

        encoder := json.NewEncoder(outTarget)
        encoder.SetIndent("", "  ")
        if err := encoder.Encode(result); err != nil {
                return err
        }

        return nil
}
