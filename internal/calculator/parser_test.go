package calculator

import (
	"reflect"
	"testing"
)

func TestOperators(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		want       []Operator
	}{
		{
			name:       "простое сложение",
			expression: "2+2",
			want: []Operator{
				{Type: "number", Value: "2"},
				{Type: "operator", Value: "+"},
				{Type: "number", Value: "2"},
			},
		},
		{
			name:       "простое сложение с пробелами",
			expression: "2 + 2",
			want: []Operator{
				{Type: "number", Value: "2"},
				{Type: "operator", Value: "+"},
				{Type: "number", Value: "2"},
			},
		},
		{
			name:       "комплексное выражение",
			expression: "2+2*3",
			want: []Operator{
				{Type: "number", Value: "2"},
				{Type: "operator", Value: "+"},
				{Type: "number", Value: "2"},
				{Type: "operator", Value: "*"},
				{Type: "number", Value: "3"},
			},
		},
		{
			name:       "выражение с дробными числами",
			expression: "2.5+3.5",
			want: []Operator{
				{Type: "number", Value: "2.5"},
				{Type: "operator", Value: "+"},
				{Type: "number", Value: "3.5"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := operators(tt.expression)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ошибка: получено %v, ожидалось %v", got, tt.want)
			}
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name             string
		expression       string
		expectedOperator string // Оператор  который   должен присутствовать   в полученых операциях
		wantErr          bool
	}{
		{
			name:             "простое сложение",
			expression:       "2+3",
			expectedOperator: "+",
			wantErr:          false,
		},
		{
			name:             "простое вычитание",
			expression:       "2*3",
			expectedOperator: "*",
			wantErr:          false,
		},
		{
			name:             "простое умножение",
			expression:       "2+3*4",
			expectedOperator: "+",
			wantErr:          false,
		},
		{
			name:             "комплексное выражение",
			expression:       "2+3*4+5",
			expectedOperator: "+",
			wantErr:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.expression)
			got, err := p.Parse()

			if (err != nil) != tt.wantErr {
				t.Errorf("ошибка: получено %v, ожидалось %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Проверяем, что операции были получены
				if len(got) == 0 {
					t.Errorf("ошибка: операции не были получены")
					return
				}

				// Выводим операции для отладки
				t.Logf("полученные операции: %v", got)

				// Проверяем, что ожидаемый оператор присутствует
				foundExpectedOperator := false
				for _, op := range got {
					if op.Operator == tt.expectedOperator {
						foundExpectedOperator = true
						break
					}
				}

				if !foundExpectedOperator {
					t.Errorf("выражение не содержит оператора %s", tt.expectedOperator)
				}
			}
		})
	}
}

func TestHelperFunctions(t *testing.T) {
	// Тест функции isDigit
	digits := []byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}
	for _, d := range digits {
		if !isDigit(d) {
			t.Errorf("число %c не распознано как цифра", d)
		}
	}

	nonDigits := []byte{'a', 'z', '+', '-', ' ', '.'}
	for _, d := range nonDigits {
		if d != '.' && isDigit(d) {
			t.Errorf("число %c распознано как цифра", d)
		}
	}

	// Тест функции isOperator
	operators := []byte{'+', '-', '*', '/'}
	for _, op := range operators {
		if !isOperator(op) {
			t.Errorf("число %c не распознано как оператор", op)
		}
	}

	nonOperators := []byte{'a', 'z', '0', '9', ' ', '.'}
	for _, op := range nonOperators {
		if isOperator(op) {
			t.Errorf("число %c распознано как оператор", op)
		}
	}
}
