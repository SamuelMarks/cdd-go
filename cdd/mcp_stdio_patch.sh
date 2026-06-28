#!/bin/bash
sed -i '' 's/run(\[\]string{"from_openapi", "to_sdk", "-i", in, "-o", out})/FromOpenAPI("to_sdk", in, out, false, false, false, false)/g' cdd/mcp_stdio.go
sed -i '' 's/run(\[\]string{"to_openapi", "-i", in, "-o", out})/ToOpenAPI(in, out)/g' cdd/mcp_stdio.go
