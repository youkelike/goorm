package reflect

import "reflect"

func IterateFunc(entity any) (map[string]FuncInfo, error) {
	typ := reflect.TypeOf(entity)
	numMethod := typ.NumMethod()
	res := make(map[string]FuncInfo, numMethod)
	for i := 0; i < numMethod; i++ {
		method := typ.Method(i)

		numIn := method.Func.Type().NumIn()
		input := make([]reflect.Type, 0, numIn)
		inputValues := make([]reflect.Value, 0, numIn)

		// 绑定参数处理
		input = append(input, reflect.TypeOf(entity))
		inputValues = append(inputValues, reflect.ValueOf(entity))

		// 正常参数处理
		for j := 1; j < numIn; j++ {
			fnInTyp := method.Func.Type().In(j)
			input = append(input, fnInTyp)
			inputValues = append(inputValues, reflect.Zero(fnInTyp))
		}

		numOut := method.Func.Type().NumOut()
		output := make([]reflect.Type, 0, numOut)
		for j := 0; j < numOut; j++ {
			output = append(output, method.Func.Type().Out(j))
		}

		resValues := method.Func.Call(inputValues)
		result := make([]any, 0, len(resValues))
		for _, v := range resValues {
			result = append(result, v.Interface())
		}

		res[method.Name] = FuncInfo{
			Name:        method.Name,
			InputTypes:  input,
			OutputTypes: output,
			Result:      result,
		}
	}
	return res, nil
}

type FuncInfo struct {
	Name        string
	InputTypes  []reflect.Type
	OutputTypes []reflect.Type
	Result      []any
}
