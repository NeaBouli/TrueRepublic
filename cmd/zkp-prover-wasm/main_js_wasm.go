//go:build js && wasm

// Command zkp-prover-wasm exposes the isolated test-only Groth16 prover to a
// maintained-client compatibility harness. It does not contain transaction,
// RPC, wallet, or broadcast functionality.
package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"

	"truerepublic/internal/zkpprover"
)

const globalFunction = "trueRepublicTestOnlyGroth16Prove"

func main() {
	function := js.FuncOf(prove)
	js.Global().Set(globalFunction, function)
	select {}
}

func prove(_ js.Value, arguments []js.Value) any {
	if len(arguments) != 4 || arguments[0].Type() != js.TypeString {
		return failure("expected request JSON and exact CS, PK, and VK Uint8Arrays")
	}
	cs, err := copyExactBytes("constraint system", arguments[1], zkpprover.ConstraintSystemSize)
	if err != nil {
		return failure(err.Error())
	}
	pk, err := copyExactBytes("proving key", arguments[2], zkpprover.ProvingKeySize)
	if err != nil {
		return failure(err.Error())
	}
	vk, err := copyExactBytes("verifying key", arguments[3], zkpprover.VerifyingKeySize)
	if err != nil {
		return failure(err.Error())
	}
	requestBytes := []byte(arguments[0].String())
	defer clear(requestBytes)
	request, err := zkpprover.DecodeRequestStrict(requestBytes)
	if err != nil {
		return failure(err.Error())
	}
	result, err := zkpprover.Prove(cs, pk, vk, request)
	if err != nil {
		return failure(err.Error())
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		return failure(fmt.Sprintf("encode result: %v", err))
	}
	return map[string]any{"ok": true, "result": string(encoded), "error": ""}
}

func copyExactBytes(label string, value js.Value, size int) ([]byte, error) {
	uint8Array := js.Global().Get("Uint8Array")
	if value.Type() != js.TypeObject || !value.InstanceOf(uint8Array) {
		return nil, fmt.Errorf("%s must be a Uint8Array", label)
	}
	if value.Get("byteLength").Int() != size || value.Get("length").Int() != size {
		return nil, fmt.Errorf("%s size = %d, want %d", label, value.Get("byteLength").Int(), size)
	}
	result := make([]byte, size)
	if copied := js.CopyBytesToGo(result, value); copied != size {
		return nil, fmt.Errorf("copy %s: copied %d bytes, want %d", label, copied, size)
	}
	return result, nil
}

func failure(message string) map[string]any {
	return map[string]any{"ok": false, "result": "", "error": message}
}

func clear(data []byte) {
	for index := range data {
		data[index] = 0
	}
}
