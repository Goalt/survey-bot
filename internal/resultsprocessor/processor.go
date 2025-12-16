package resultsprocessor

import (
	"fmt"

	"git.ykonkov.com/ykonkov/survey-bot/internal/entity"
)

var (
	calculationsType = map[string]func(entity.Survey, []entity.Answer) entity.Results{
		"test_1": calculateTest1,
		"test_2": calculateTest2,
		"test_3": calculateTest3,
		"test_4": calculateTest4,
		"test_5": calculateTest5,
		"test_6": calculateTest6,
	}
)

type processor struct{}

func New() *processor {
	return &processor{}
}

func (p *processor) GetResults(survey entity.Survey, answers []entity.Answer) (entity.Results, error) {
	f, ok := calculationsType[survey.CalculationsType]
	if !ok {
		return entity.Results{}, fmt.Errorf("unknown calculations type: %s", survey.CalculationsType)
	}

	return f(survey, answers), nil
}

// Check questions for correctness
// Check if it could be processed (computation key should exist)
func (p *processor) Validate(survey entity.Survey) error {
	if err := survey.Validate(); err != nil {
		return fmt.Errorf("failed to validate survey, %w", err)
	}

	if _, ok := calculationsType[survey.CalculationsType]; !ok {
		return fmt.Errorf("unknown calculations type: %s", survey.CalculationsType)
	}

	return nil
}

func calculateTest1(survey entity.Survey, answers []entity.Answer) entity.Results {
	var (
		// Эмоциональное истощение
		s1 int
		// Деперсонализация
		s2 int
		// Редукция профессионализма
		s3 int
	)

	get := func(id int) int {
		return answers[id-1].Data[0] - 1
	}

	// 1, 2, 3, 8, 13, 14, 16, 20
	s1 = get(1) + get(2) + get(3) + get(8) + get(13) + get(14) + get(16) + get(20) - get(6)

	// 5, 10, 11, 15, 22
	s2 = get(5) + get(10) + get(11) + get(15) + get(22)

	// 4, 7, 9, 12, 17, 18, 19, 21
	s3 = get(4) + get(7) + get(9) + get(12) + get(17) + get(18) + get(19) + get(21)

	var (
		s1Level string
		s2Level string
		s3Level string
	)

	switch {
	case s1 <= 15:
		s1Level = "низкий уровень"
	case s1 > 15 && s1 <= 24:
		s1Level = "средний уровень"
	default:
		s1Level = "высокий уровень"
	}

	switch {
	case s2 <= 5:
		s2Level = "низкий уровень"
	case s2 > 5 && s2 <= 10:
		s2Level = "средний уровень"
	default:
		s2Level = "высокий уровень"
	}

	switch {
	case s3 >= 37:
		s3Level = "низкий уровень"
	case s3 >= 31 && s3 <= 36:
		s3Level = "средний уровень"
	default:
		s3Level = "высокий уровень"
	}

	result := fmt.Sprintf("Эмоциональное истощение - %s, Деперсонализация - %s, Редукция профессионализма - %s", s1Level, s2Level, s3Level)

	return entity.Results{
		Text: result,
		Metadata: entity.ResultsMetadata{
			Raw: map[string]interface{}{
				"s1": s1,
				"s2": s2,
				"s3": s3,
			},
		},
	}
}

func calculateTest2(survey entity.Survey, answers []entity.Answer) entity.Results {
	var coef = [][]float64{
		{
			0,
			0.034,
			0.041,
			0.071,
			0.458,
		},
		{
			0,
			0.062,
			0.075,
			0.117,
			0.246,
		},
		{
			0,
			0.059,
			0.073,
			0.129,
			0.242,
		},
		{
			0,
			0.053,
			0.066,
			0.19,
			0.377,
		},
		{
			0,
			0.033,
			0.041,
			0.109,
			0.179,
		},
		{
			0,
			0,
			0,
			0,
			0,
		},
	}

	var s float64

	for i, answer := range answers {
		if len(answers)-1 == i {
			continue
		}

		tmp := coef[i][answer.Data[0]-1] * float64(answer.Data[0])
		s += tmp
	}

	s = 1 - s

	return entity.Results{
		Text: fmt.Sprintf("Сумма баллов: %.2f", s),
		Metadata: entity.ResultsMetadata{
			Raw: map[string]interface{}{
				"s": s,
			},
		},
	}
}

func calculateTest3(survey entity.Survey, answers []entity.Answer) entity.Results {
	var (
		// РЕАКТИВНАЯ ТРЕВОЖНОСТЬ
		s1 int
		// ЛИЧНОСТНАЯ ТРЕВОЖНОСТЬ
		s2 int
	)

	get := func(id int) int {
		return answers[id-1].Data[0] - 1
	}

	// 3,4,6,7,9,12,13,14,17,18
	s1 = get(3) + get(4) + get(6) + get(7) + get(9) + get(12) + get(13) + get(14) + get(17) + get(18)
	// 1,2,5,8,10,11,15,16,19,20.
	s1 = s1 - (get(1) + get(2) + get(5) + get(8) + get(10) + get(11) + get(15) + get(16) + get(19) + get(20)) + 50

	// 22,23,24,25,28,29,31,32,34,35,37,38,40
	s2 = get(22) + get(23) + get(24) + get(25) + get(28) + get(29) + get(31) + get(32) + get(34) + get(35) + get(37) + get(38) + get(40)
	// 21,26,27,30,33,36,39
	s2 = s2 - (get(21) + get(26) + get(27) + get(30) + get(33) + get(36) + get(39)) + 35

	var (
		s1Level string
		s2Level string
	)

	switch {
	case s1 <= 30:
		s1Level = "низкий уровень"
	case s1 > 30 && s1 <= 45:
		s1Level = "средний уровень"
	default:
		s1Level = "высокий уровень"
	}

	switch {
	case s2 <= 30:
		s2Level = "низкий уровень"
	case s2 > 30 && s2 <= 45:
		s2Level = "средний уровень"
	default:
		s2Level = "высокий уровень"
	}

	result := fmt.Sprintf("РЕАКТИВНАЯ ТРЕВОЖНОСТЬ - %s, ЛИЧНОСТНАЯ ТРЕВОЖНОСТЬ - %s", s1Level, s2Level)

	return entity.Results{
		Text: result,
		Metadata: entity.ResultsMetadata{
			Raw: map[string]interface{}{
				"s1": s1,
				"s2": s2,
			},
		},
	}
}

func calculateTest4(survey entity.Survey, answers []entity.Answer) entity.Results {
	var (
		s       int
		s1Level string
	)

	for i := 0; i < 18; i++ {
		s += answers[i].Data[0]
	}

	if answers[19].Data[0] == 1 {
		s += answers[18].Data[0]
	}

	s += answers[20].Data[0]
	s += answers[21].Data[0]

	switch {
	case s <= 9:
		s1Level = "отсутствие депрессивных симптомов"
	case s > 9 && s <= 15:
		s1Level = "легкая депрессия (субдепрессия)"
	case s > 15 && s <= 19:
		s1Level = "умеренная депрессия"
	case s > 19 && s <= 29:
		s1Level = "выраженная депрессия (средней тяжести)"
	default:
		s1Level = "тяжелая депрессия"
	}

	return entity.Results{
		Text: s1Level,
		Metadata: entity.ResultsMetadata{
			Raw: map[string]interface{}{
				"s": s,
			},
		},
	}
}

func calculateTest5(survey entity.Survey, answers []entity.Answer) entity.Results {
	var (
		// Экстраверсия - интроверсия
		s1 int
		// Нейротизм
		s2 int
		// Шкала лжи
		s3 int
	)

	get := func(id int, expected int) int {
		if answers[id-1].Data[0] == expected {
			return 1
		}

		return 0
	}

	s1 = get(1, 1) +
		get(3, 1) +
		get(8, 1) +
		get(10, 1) +
		get(13, 1) +
		get(17, 1) +
		get(22, 1) +
		get(25, 1) +
		get(27, 1) +
		get(39, 1) +
		get(44, 1) +
		get(46, 1) +
		get(49, 1) +
		get(53, 1) +
		get(56, 1) +
		get(5, 2) +
		get(15, 2) +
		get(20, 2) +
		get(29, 2) +
		get(32, 2) +
		get(34, 2) +
		get(37, 2) +
		get(41, 2) +
		get(51, 2)

	s2 = get(2, 1) +
		get(4, 1) +
		get(7, 1) +
		get(9, 1) +
		get(11, 1) +
		get(14, 1) +
		get(16, 1) +
		get(19, 1) +
		get(21, 1) +
		get(23, 1) +
		get(26, 1) +
		get(28, 1) +
		get(31, 1) +
		get(33, 1) +
		get(35, 1) +
		get(38, 1) +
		get(40, 1) +
		get(43, 1) +
		get(45, 1) +
		get(47, 1) +
		get(50, 1) +
		get(52, 1) +
		get(55, 1) +
		get(57, 1)

	s3 = get(6, 1) +
		get(24, 1) +
		get(36, 1) +
		get(12, 2) +
		get(18, 2) +
		get(30, 2) +
		get(42, 2) +
		get(48, 2) +
		get(54, 2)

	var (
		s1Level string
		s2Level string
		s3Level string
	)

	switch {
	case s1 > 19:
		s1Level = "яркий экстраверт"
	case s1 > 15:
		s1Level = "экстраверт"
	case s1 > 9:
		s1Level = "норма"
	case s1 > 5:
		s1Level = "интроверт"
	default:
		s1Level = "глубокий интроверт"
	}

	switch {
	case s2 > 19:
		s2Level = "очень высокий уровень нейротизма"
	case s2 > 14:
		s2Level = "высокий уровень нейротизма"
	case s2 > 9:
		s2Level = "среднее значение"
	default:
		s2Level = "низкий уровень нейротизма"
	}

	switch {
	case s3 > 4:
		s3Level = "неискренность в ответах"
	default:
		s3Level = "норма"
	}

	var description = `Интроверт это человек, психический склад которого характеризуется сосредоточенностью на своем внутреннем мире, замкнутостью, созерцательностью; тот, кто не склонен к общению и с трудом устанавливает контакты с окружающим миром
Экстраверт это общительный, экспрессивный человек с активной социальной позицией. Его переживания и интересы направлены на внешний мир. Экстраверты удовлетворяют большинство своих потребностей через взаимодействие с людьми.
Нейротизм – это личностная черта человека, которая проявляется в беспокойстве, тревожности и эмоциональной неустойчивости. Нейротизм в психологии это индивидуальная переменная, которая выражает особенности нервной системы (лабильность и реактивность). Те люди, у которых высокий уровень нейротизма, под внешним выражением полного благополучия скрывают внутреннюю неудовлетворенность и личные конфликты. Они реагируют на всё происходящие чересчур эмоционально и не всегда адекватно к ситуации.`

	result := fmt.Sprintf(`"Экстраверсия - интроверсия" - %s, "Нейротизм" - %s, "Шкала лжи" - %s

%s`, s1Level, s2Level, s3Level, description)

	return entity.Results{
		Text: result,
		Metadata: entity.ResultsMetadata{
			Raw: map[string]interface{}{
				"estraversia-introversia": s1,
				"neurotism":               s2,
				"lie":                     s3,
			},
		},
	}
}

func calculateTest6(survey entity.Survey, answers []entity.Answer) entity.Results {
	var (
		// Реалистический тип
		s1 int
		// Интеллектуальный тип
		s2 int
		// Социальный тип
		s3 int
		// Конвенциальный тип
		s4 int
		// Предприимчивый тип
		s5 int
		// Артистический тип
		s6 int
	)

	get := func(id int, expected int) int {
		if answers[id-1].Data[0] == expected {
			return 1
		}

		return 0
	}

	// 1а, 2а, 3а, 4а, 5а, 16а, 17а, 18а, 19а, 21а, 31а, 32а, 33а, 34а
	s1 = get(1, 1) +
		get(2, 1) +
		get(3, 1) +
		get(4, 1) +
		get(5, 1) +
		get(16, 1) +
		get(17, 1) +
		get(18, 1) +
		get(19, 1) +
		get(21, 1) +
		get(31, 1) +
		get(32, 1) +
		get(33, 1) +
		get(34, 1)

	// 1б, 6а, 7а, 8а, 9а, 16б, 20а, 22а, 23а, 24а, 31б, 35а, 36а, 37а.
	s2 = get(1, 2) +
		get(6, 1) +
		get(7, 1) +
		get(8, 1) +
		get(9, 1) +
		get(16, 2) +
		get(20, 1) +
		get(22, 1) +
		get(23, 1) +
		get(24, 1) +
		get(31, 2) +
		get(35, 1) +
		get(36, 1) +
		get(37, 1)

	// 2б, 6б, 10а, 11а, 12а, 17б, 29б, 25а, 26а, 27а, 36б, 38а, 39а, 41б.
	s3 = get(2, 2) +
		get(6, 2) +
		get(10, 1) +
		get(11, 1) +
		get(12, 1) +
		get(17, 2) +
		get(29, 2) +
		get(25, 1) +
		get(26, 1) +
		get(27, 1) +
		get(36, 2) +
		get(38, 1) +
		get(39, 1) +
		get(41, 2)

	// 3б, 7б, 10б, 13а, 14а, 18б, 22б, 25б, 28а, 29а, 32б, 38б, 40а, 42а.
	s4 = get(3, 2) +
		get(7, 2) +
		get(10, 2) +
		get(13, 1) +
		get(14, 1) +
		get(18, 2) +
		get(22, 2) +
		get(25, 2) +
		get(28, 1) +
		get(29, 1) +
		get(32, 2) +
		get(38, 2) +
		get(40, 1) +
		get(42, 1)

	// 4б, 8б, 11б, 13б, 15а, 23б, 28б, 30а, 33б, 35б, 37б, 39б, 40б.
	s5 = get(4, 2) +
		get(8, 2) +
		get(11, 2) +
		get(13, 2) +
		get(15, 1) +
		get(23, 2) +
		get(28, 2) +
		get(30, 1) +
		get(33, 2) +
		get(35, 2) +
		get(37, 2) +
		get(39, 2) +
		get(40, 2)

	// 5б, 9б, 12б, 14б, 15б, 19б, 21б, 24а, 27б, 29б, 30б, 34б, 41а, 42б.
	s6 = get(5, 2) +
		get(9, 2) +
		get(12, 2) +
		get(14, 2) +
		get(15, 2) +
		get(19, 2) +
		get(21, 2) +
		get(24, 1) +
		get(27, 2) +
		get(29, 2) +
		get(30, 2) +
		get(34, 2) +
		get(41, 1) +
		get(42, 2)

	var description = `Реалистический тип – этому типу личности свойственна эмоциональная стабильность, ориентация на настоящее. Представители данного типа занимаются конкретными объектами и их практическим использованием: вещами, инструментами, машинами. Отдают предпочтение занятиям требующим моторных навыков, ловкости, конкретности.
Интеллектуальный тип – ориентирован на умственный труд. Он аналитичен, рационален, независим, оригинален. Преобладают теоретические и в некоторой степени эстетические ценности. Размышления о проблеме он предпочитает занятиям по реализации связанных с ней решений. Ему нравится решать задачи, требующие абстрактного мышления.
Социальный тип - ставит перед собой такие цели и задачи, которые позволяют им установить тесный контакт с окружающей социальной средой. Обладает социальными умениями и нуждается в социальных контактах. Стремятся поучать, воспитывать. Гуманны. Способны приспособиться практически к любым условиям. Стараются держаться в стороне от интеллектуальных проблем. Они активны и решают проблемы, опираясь главным образом на эмоции, чувства и умение общаться.
Конвенциальный тип – отдает предпочтение четко структурированной деятельности. Из окружающей его среды он выбирает цели, задачи и ценности, проистекающие из обычаев и обусловленные состоянием общества. Ему характерны серьезность настойчивость, консерватизм, исполнительность. В соответствии с этим его подход к проблемам носит стереотипичный, практический и конкретный характер.
Предприимчивый тип – избирает цели, ценности и задачи, позволяющие ему проявить энергию, энтузиазм, импульсивность, доминантность, реализовать любовь к приключенчеству. Ему не по душе занятия, связанные с ручным трудом, а также требующие усидчивости, большой концентрации внимания и интеллектуальных усилий. Предпочитает руководящие роли в которых может удовлетворять свои потребности в доминантности и признании. Активен, предприимчив.
Артистический тип – отстраняется от отчетливо структурированных проблем и видов деятельности, предполагающих большую физическую силу. В общении с окружающими опираются на свои непосредственные ощущения, эмоции, интуицию и воображение. Ему присущ сложный взгляд на жизнь, гибкость, независимость суждений. Свойственна несоциальность, оригинальность.`

	result := fmt.Sprintf("Реалистический тип - %d, Интеллектуальный тип - %d, Социальный тип - %d, Конвенциальный тип - %d, Предприимчивый тип - %d, Артистический тип - %d\n\n%s", s1, s2, s3, s4, s5, s6, description)

	return entity.Results{
		Text: result,
		Metadata: entity.ResultsMetadata{
			Raw: map[string]interface{}{
				"realistic":    s1,
				"intillectual": s2,
				"social":       s3,
				"conventional": s4,
				"enterprising": s5,
				"artistic":     s6,
			},
		},
	}
}
