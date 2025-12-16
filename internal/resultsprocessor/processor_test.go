package resultsprocessor

import (
	"reflect"
	"testing"

	"git.ykonkov.com/ykonkov/survey-bot/internal/entity"
	"git.ykonkov.com/ykonkov/survey-bot/internal/service"
	"github.com/stretchr/testify/require"
)

func Test_calculateTest1(t *testing.T) {
	test1, err := service.ReadSurveyFromFile("../../surveytests/1.json")
	require.NoError(t, err)

	type args struct {
		survey  entity.Survey
		answers []entity.Answer
	}
	tests := []struct {
		name string
		args args
		want entity.Results
	}{
		{
			name: "test1",
			args: args{
				survey:  test1,
				answers: generateSelectAnswersWithStep(1, 2, 1, 1, 4, 2, 3, 4, 1, 4, 2, 1, 5, 1, 2, 2, 3, 5, 4, 3, 2, 4, 2),
			},
			want: entity.Results{
				Text: "Эмоциональное истощение - низкий уровень, Деперсонализация - средний уровень, Редукция профессионализма - средний уровень",
				Metadata: entity.ResultsMetadata{
					Raw: map[string]interface{}{
						"s1": 10,
						"s2": 9,
						"s3": 33,
					},
				},
			},
		},
		{
			name: "test1",
			args: args{
				survey:  test1,
				answers: generateSelectAnswersWithStep(1, 1, 3, 1, 6, 1, 0, 5, 1, 5, 0, 0, 6, 0, 0, 1, 0, 5, 3, 6, 0, 3, 0),
			},
			want: entity.Results{
				Text: "Эмоциональное истощение - низкий уровень, Деперсонализация - низкий уровень, Редукция профессионализма - низкий уровень",
				Metadata: entity.ResultsMetadata{
					Raw: map[string]interface{}{
						"s1": 6,
						"s2": 2,
						"s3": 39,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateTest1(tt.args.survey, tt.args.answers); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("calculateTest1() = %v, want %v", got, tt.want)
			}
		})
	}
}

func generateSelectAnswersWithStep(step int, answers ...int) []entity.Answer {
	var result []entity.Answer
	for _, answer := range answers {
		result = append(result, entity.Answer{Type: entity.AnswerTypeSelect, Data: []int{answer + step}})
	}
	return result
}

func Test_calculateTest4(t *testing.T) {
	test4, err := service.ReadSurveyFromFile("../../surveytests/4.json")
	require.NoError(t, err)

	type args struct {
		survey  entity.Survey
		answers []entity.Answer
	}
	tests := []struct {
		name string
		args args
		want entity.Results
	}{
		{
			name: "test4",
			args: args{
				survey: test4,
				answers: append(
					generateSelectAnswersWithStep(0, 0, 1, 0, 0, 0, 1, 0, 1, 0, 0, 1, 0, 0, 0, 0, 0, 1, 0, 0),
					entity.Answer{Type: entity.AnswerTypeSelect, Data: []int{1}},
					entity.Answer{Type: entity.AnswerTypeSelect, Data: []int{0}},
					entity.Answer{Type: entity.AnswerTypeSelect, Data: []int{0}},
				),
			},
			want: entity.Results{
				Text: "отсутствие депрессивных симптомов",
				Metadata: entity.ResultsMetadata{
					Raw: map[string]interface{}{
						"s": 5,
					},
				},
			},
		},
		{
			name: "test4",
			args: args{
				survey: test4,
				answers: append(
					generateSelectAnswersWithStep(0, 1, 1, 2, 1, 1, 0, 0, 1, 0, 0, 0, 1, 1, 1, 0, 0, 0, 0, 0),
					entity.Answer{Type: entity.AnswerTypeSelect, Data: []int{1}},
					entity.Answer{Type: entity.AnswerTypeSelect, Data: []int{1}},
					entity.Answer{Type: entity.AnswerTypeSelect, Data: []int{1}},
				),
			},
			want: entity.Results{
				Text: "легкая депрессия (субдепрессия)",
				Metadata: entity.ResultsMetadata{
					Raw: map[string]interface{}{
						"s": 12,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateTest4(tt.args.survey, tt.args.answers); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("calculateTest4() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_calculateTest3(t *testing.T) {
	test3, err := service.ReadSurveyFromFile("../../surveytests/3.json")
	require.NoError(t, err)

	type args struct {
		survey  entity.Survey
		answers []entity.Answer
	}
	tests := []struct {
		name string
		args args
		want entity.Results
	}{
		{
			name: "test3",
			args: args{
				survey:  test3,
				answers: generateSelectAnswersWithStep(1, 3, 3, 1, 1, 3, 1, 1, 2, 1, 3, 3, 1, 1, 1, 3, 3, 1, 1, 2, 3, 3, 2, 1, 2, 2, 3, 3, 1, 2, 3, 2, 1, 3, 2, 1, 3, 2, 1, 3, 2),
			},
			want: entity.Results{
				Text: "РЕАКТИВНАЯ ТРЕВОЖНОСТЬ - средний уровень, ЛИЧНОСТНАЯ ТРЕВОЖНОСТЬ - средний уровень",
				Metadata: entity.ResultsMetadata{
					Raw: map[string]interface{}{
						"s1": 32,
						"s2": 35,
					},
				},
			},
		},
		{
			name: "test3",
			args: args{
				survey:  test3,
				answers: generateSelectAnswersWithStep(1, 1, 1, 2, 1, 1, 2, 1, 1, 2, 2, 2, 2, 1, 1, 2, 1, 1, 1, 1, 1, 2, 2, 2, 2, 1, 2, 2, 2, 2, 2, 3, 1, 3, 2, 2, 2, 2, 3, 2, 2),
			},
			want: entity.Results{
				Text: "РЕАКТИВНАЯ ТРЕВОЖНОСТЬ - высокий уровень, ЛИЧНОСТНАЯ ТРЕВОЖНОСТЬ - высокий уровень",
				Metadata: entity.ResultsMetadata{
					Raw: map[string]interface{}{
						"s1": 51,
						"s2": 46,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateTest3(tt.args.survey, tt.args.answers); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("calculateTest3() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_calculateTest2(t *testing.T) {
	test2, err := service.ReadSurveyFromFile("../../surveytests/2.json")
	require.NoError(t, err)

	type args struct {
		survey  entity.Survey
		answers []entity.Answer
	}
	tests := []struct {
		name string
		args args
		want entity.Results
	}{
		{
			name: "test2",
			args: args{
				survey:  test2,
				answers: generateSelectAnswersWithStep(0, 1, 1, 2, 2, 2, 2, 99),
			},
			want: entity.Results{
				Text: "Сумма баллов: 0.71",
				Metadata: entity.ResultsMetadata{
					Raw: map[string]interface{}{
						"s": 0.71,
					},
				},
			},
		},
		{
			name: "test2",
			args: args{
				survey:  test2,
				answers: generateSelectAnswersWithStep(0, 2, 3, 3, 2, 3, 2, 99),
			},
			want: entity.Results{
				Text: "Сумма баллов: 0.26",
				Metadata: entity.ResultsMetadata{
					Raw: map[string]interface{}{
						"s": 0.259,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateTest2(tt.args.survey, tt.args.answers); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("calculateTest2() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_calculateTest5(t *testing.T) {
	test5, err := service.ReadSurveyFromFile("../../surveytests/5.json")
	require.NoError(t, err)

	type args struct {
		survey  entity.Survey
		answers []entity.Answer
	}
	tests := []struct {
		name string
		args args
		want entity.Results
	}{
		{
			name: "test5_simple1",
			args: args{
				survey:  test5,
				answers: generateSelectAnswersWithStep(0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1),
			},
			want: entity.Results{
				Text: `"Экстраверсия - интроверсия" - норма, "Нейротизм" - очень высокий уровень нейротизма, "Шкала лжи" - норма

Интроверт это человек, психический склад которого характеризуется сосредоточенностью на своем внутреннем мире, замкнутостью, созерцательностью; тот, кто не склонен к общению и с трудом устанавливает контакты с окружающим миром
Экстраверт это общительный, экспрессивный человек с активной социальной позицией. Его переживания и интересы направлены на внешний мир. Экстраверты удовлетворяют большинство своих потребностей через взаимодействие с людьми.
Нейротизм – это личностная черта человека, которая проявляется в беспокойстве, тревожности и эмоциональной неустойчивости. Нейротизм в психологии это индивидуальная переменная, которая выражает особенности нервной системы (лабильность и реактивность). Те люди, у которых высокий уровень нейротизма, под внешним выражением полного благополучия скрывают внутреннюю неудовлетворенность и личные конфликты. Они реагируют на всё происходящие чересчур эмоционально и не всегда адекватно к ситуации.`,
				Metadata: entity.ResultsMetadata{
					Raw: map[string]interface{}{
						"estraversia-introversia": 15,
						"neurotism":               24,
						"lie":                     3,
					},
				},
			},
		},
		{
			name: "test5_simple2",
			args: args{
				survey:  test5,
				answers: generateSelectAnswersWithStep(0, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2),
			},
			want: entity.Results{
				Text: `"Экстраверсия - интроверсия" - интроверт, "Нейротизм" - низкий уровень нейротизма, "Шкала лжи" - неискренность в ответах

Интроверт это человек, психический склад которого характеризуется сосредоточенностью на своем внутреннем мире, замкнутостью, созерцательностью; тот, кто не склонен к общению и с трудом устанавливает контакты с окружающим миром
Экстраверт это общительный, экспрессивный человек с активной социальной позицией. Его переживания и интересы направлены на внешний мир. Экстраверты удовлетворяют большинство своих потребностей через взаимодействие с людьми.
Нейротизм – это личностная черта человека, которая проявляется в беспокойстве, тревожности и эмоциональной неустойчивости. Нейротизм в психологии это индивидуальная переменная, которая выражает особенности нервной системы (лабильность и реактивность). Те люди, у которых высокий уровень нейротизма, под внешним выражением полного благополучия скрывают внутреннюю неудовлетворенность и личные конфликты. Они реагируют на всё происходящие чересчур эмоционально и не всегда адекватно к ситуации.`,
				Metadata: entity.ResultsMetadata{
					Raw: map[string]interface{}{
						"estraversia-introversia": 9,
						"neurotism":               0,
						"lie":                     6,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateTest5(tt.args.survey, tt.args.answers); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("calculateTest5() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_calculateTest6(t *testing.T) {
	test5, err := service.ReadSurveyFromFile("../../surveytests/6.json")
	require.NoError(t, err)

	type args struct {
		survey  entity.Survey
		answers []entity.Answer
	}
	tests := []struct {
		name string
		args args
		want entity.Results
	}{
		{
			name: "test6_simple1",
			args: args{
				survey:  test5,
				answers: generateSelectAnswersWithStep(0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1),
			},
			want: entity.Results{
				Text: `Реалистический тип - 14, Интеллектуальный тип - 11, Социальный тип - 8, Конвенциальный тип - 6, Предприимчивый тип - 2, Артистический тип - 2

Реалистический тип – этому типу личности свойственна эмоциональная стабильность, ориентация на настоящее. Представители данного типа занимаются конкретными объектами и их практическим использованием: вещами, инструментами, машинами. Отдают предпочтение занятиям требующим моторных навыков, ловкости, конкретности.
Интеллектуальный тип – ориентирован на умственный труд. Он аналитичен, рационален, независим, оригинален. Преобладают теоретические и в некоторой степени эстетические ценности. Размышления о проблеме он предпочитает занятиям по реализации связанных с ней решений. Ему нравится решать задачи, требующие абстрактного мышления.
Социальный тип - ставит перед собой такие цели и задачи, которые позволяют им установить тесный контакт с окружающей социальной средой. Обладает социальными умениями и нуждается в социальных контактах. Стремятся поучать, воспитывать. Гуманны. Способны приспособиться практически к любым условиям. Стараются держаться в стороне от интеллектуальных проблем. Они активны и решают проблемы, опираясь главным образом на эмоции, чувства и умение общаться.
Конвенциальный тип – отдает предпочтение четко структурированной деятельности. Из окружающей его среды он выбирает цели, задачи и ценности, проистекающие из обычаев и обусловленные состоянием общества. Ему характерны серьезность настойчивость, консерватизм, исполнительность. В соответствии с этим его подход к проблемам носит стереотипичный, практический и конкретный характер.
Предприимчивый тип – избирает цели, ценности и задачи, позволяющие ему проявить энергию, энтузиазм, импульсивность, доминантность, реализовать любовь к приключенчеству. Ему не по душе занятия, связанные с ручным трудом, а также требующие усидчивости, большой концентрации внимания и интеллектуальных усилий. Предпочитает руководящие роли в которых может удовлетворять свои потребности в доминантности и признании. Активен, предприимчив.
Артистический тип – отстраняется от отчетливо структурированных проблем и видов деятельности, предполагающих большую физическую силу. В общении с окружающими опираются на свои непосредственные ощущения, эмоции, интуицию и воображение. Ему присущ сложный взгляд на жизнь, гибкость, независимость суждений. Свойственна несоциальность, оригинальность.`,
				Metadata: entity.ResultsMetadata{
					Raw: map[string]interface{}{
						"realistic":    14,
						"intillectual": 11,
						"social":       8,
						"conventional": 6,
						"enterprising": 2,
						"artistic":     2,
					},
				},
			},
		},
		{
			name: "test6_simple2",
			args: args{
				survey:  test5,
				answers: generateSelectAnswersWithStep(0, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2),
			},
			want: entity.Results{
				Text: `Реалистический тип - 0, Интеллектуальный тип - 3, Социальный тип - 6, Конвенциальный тип - 8, Предприимчивый тип - 11, Артистический тип - 12

Реалистический тип – этому типу личности свойственна эмоциональная стабильность, ориентация на настоящее. Представители данного типа занимаются конкретными объектами и их практическим использованием: вещами, инструментами, машинами. Отдают предпочтение занятиям требующим моторных навыков, ловкости, конкретности.
Интеллектуальный тип – ориентирован на умственный труд. Он аналитичен, рационален, независим, оригинален. Преобладают теоретические и в некоторой степени эстетические ценности. Размышления о проблеме он предпочитает занятиям по реализации связанных с ней решений. Ему нравится решать задачи, требующие абстрактного мышления.
Социальный тип - ставит перед собой такие цели и задачи, которые позволяют им установить тесный контакт с окружающей социальной средой. Обладает социальными умениями и нуждается в социальных контактах. Стремятся поучать, воспитывать. Гуманны. Способны приспособиться практически к любым условиям. Стараются держаться в стороне от интеллектуальных проблем. Они активны и решают проблемы, опираясь главным образом на эмоции, чувства и умение общаться.
Конвенциальный тип – отдает предпочтение четко структурированной деятельности. Из окружающей его среды он выбирает цели, задачи и ценности, проистекающие из обычаев и обусловленные состоянием общества. Ему характерны серьезность настойчивость, консерватизм, исполнительность. В соответствии с этим его подход к проблемам носит стереотипичный, практический и конкретный характер.
Предприимчивый тип – избирает цели, ценности и задачи, позволяющие ему проявить энергию, энтузиазм, импульсивность, доминантность, реализовать любовь к приключенчеству. Ему не по душе занятия, связанные с ручным трудом, а также требующие усидчивости, большой концентрации внимания и интеллектуальных усилий. Предпочитает руководящие роли в которых может удовлетворять свои потребности в доминантности и признании. Активен, предприимчив.
Артистический тип – отстраняется от отчетливо структурированных проблем и видов деятельности, предполагающих большую физическую силу. В общении с окружающими опираются на свои непосредственные ощущения, эмоции, интуицию и воображение. Ему присущ сложный взгляд на жизнь, гибкость, независимость суждений. Свойственна несоциальность, оригинальность.`,
				Metadata: entity.ResultsMetadata{
					Raw: map[string]interface{}{
						"realistic":    0,
						"intillectual": 3,
						"social":       6,
						"conventional": 8,
						"enterprising": 11,
						"artistic":     12,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateTest6(tt.args.survey, tt.args.answers); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("calculateTest5() = %v, want %v", got, tt.want)
			}
		})
	}
}
