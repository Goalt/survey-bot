package entity_test

import (
	"reflect"
	"testing"

	"git.ykonkov.com/ykonkov/survey-bot/internal/entity"
)

func TestQuestion_GetAnswer(t *testing.T) {
	type fields struct {
		Text            string
		AnswerType      entity.AnswerType
		PossibleAnswers []int
	}
	type args struct {
		answerRaw string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    entity.Answer
		wantErr bool
		errMsg  string
	}{
		// Segment
		{
			name: "segment",
			fields: fields{
				Text:            "question text",
				AnswerType:      entity.AnswerTypeSegment,
				PossibleAnswers: []int{1, 3},
			},
			args: args{
				answerRaw: "2",
			},
			want: entity.Answer{
				Type: entity.AnswerTypeSegment,
				Data: []int{2},
			},
			wantErr: false,
		},
		{
			name: "segment_out_of_range",
			fields: fields{
				Text:            "question text",
				AnswerType:      entity.AnswerTypeSegment,
				PossibleAnswers: []int{1, 3},
			},
			args: args{
				answerRaw: "5",
			},
			wantErr: true,
			errMsg:  entity.ErrAnswerOutOfRange.Error(),
		},
		{
			name: "segment_out_of_range",
			fields: fields{
				Text:            "question text",
				AnswerType:      entity.AnswerTypeSegment,
				PossibleAnswers: []int{1, 3},
			},
			args: args{
				answerRaw: "0",
			},
			wantErr: true,
			errMsg:  entity.ErrAnswerOutOfRange.Error(),
		},
		{
			name: "segment_bad_value",
			fields: fields{
				Text:            "question text",
				AnswerType:      entity.AnswerTypeSegment,
				PossibleAnswers: []int{1, 3},
			},
			args: args{
				answerRaw: "abc",
			},
			wantErr: true,
			errMsg:  `failed to parse argument, strconv.ParseInt: parsing "abc": invalid syntax: can't parse answer to segment`,
		},
		{
			name: "segment_bad_value",
			fields: fields{
				Text:            "question text",
				AnswerType:      entity.AnswerTypeSegment,
				PossibleAnswers: []int{1, 3},
			},
			args: args{
				answerRaw: "2.5",
			},
			wantErr: true,
			errMsg:  `failed to parse argument, strconv.ParseInt: parsing "2.5": invalid syntax: can't parse answer to segment`,
		},
		// select
		{
			name: "select",
			fields: fields{
				Text:            "question text",
				AnswerType:      entity.AnswerTypeSelect,
				PossibleAnswers: []int{1, 2, 3, 4},
			},
			args: args{
				answerRaw: "1",
			},
			want: entity.Answer{
				Type: entity.AnswerTypeSelect,
				Data: []int{1},
			},
			wantErr: false,
		},
		{
			name: "select_not_found",
			fields: fields{
				Text:            "question text",
				AnswerType:      entity.AnswerTypeSelect,
				PossibleAnswers: []int{1, 2, 3, 4},
			},
			args: args{
				answerRaw: "5",
			},
			wantErr: true,
			errMsg:  "answer not found",
		},
		{
			name: "select_parse_error",
			fields: fields{
				Text:            "question text",
				AnswerType:      entity.AnswerTypeSelect,
				PossibleAnswers: []int{1, 2, 3, 4},
			},
			args: args{
				answerRaw: "5.5",
			},
			wantErr: true,
			errMsg:  `failed to parse argument, strconv.ParseInt: parsing "5.5": invalid syntax: can't parse answer to select`,
		},
		// multiselect
		{
			name: "multiselect",
			fields: fields{
				Text:            "question text",
				AnswerType:      entity.AnswerTypeMultiSelect,
				PossibleAnswers: []int{1, 2, 3, 4},
			},
			args: args{
				answerRaw: "1,2",
			},
			want: entity.Answer{
				Type: entity.AnswerTypeMultiSelect,
				Data: []int{1, 2},
			},
			wantErr: false,
		},
		{
			name: "multiselect_not_found",
			fields: fields{
				Text:            "question text",
				AnswerType:      entity.AnswerTypeMultiSelect,
				PossibleAnswers: []int{1, 2, 3, 4},
			},
			args: args{
				answerRaw: "1,5",
			},
			wantErr: true,
			errMsg:  entity.ErrAnswerNotFound.Error(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := entity.Question{
				Text:            tt.fields.Text,
				AnswerType:      tt.fields.AnswerType,
				PossibleAnswers: tt.fields.PossibleAnswers,
			}
			got, err := q.GetAnswer(tt.args.answerRaw)
			if (err != nil) != tt.wantErr {
				t.Errorf("Question.GetAnswer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Question.GetAnswer() = %v, want %v", got, tt.want)
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("Question.GetAnswer() error = %v, wantErrMsg %v", err, tt.errMsg)
			}
		})
	}
}

func TestQuestion_Validate(t *testing.T) {
	type fields struct {
		Text            string
		AnswerType      entity.AnswerType
		PossibleAnswers []int
		AnswersText     []string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "segment",
			fields: fields{
				Text:            "Question text.",
				AnswerType:      entity.AnswerTypeSegment,
				PossibleAnswers: []int{1, 3},
			},
			wantErr: false,
		},
		{
			name: "select",
			fields: fields{
				Text:            "Question text.",
				AnswerType:      entity.AnswerTypeSelect,
				PossibleAnswers: []int{1, 2},
				AnswersText:     []string{"a", "b"},
			},
			wantErr: false,
		},
		{
			name: "multiselect",
			fields: fields{
				Text:            "Question text.",
				AnswerType:      entity.AnswerTypeMultiSelect,
				PossibleAnswers: []int{1, 2},
				AnswersText:     []string{"a", "b"},
			},
			wantErr: false,
		},
		{
			name: "fail, wrong answer type",
			fields: fields{
				Text:            "Question text.",
				AnswerType:      entity.AnswerType("123123"),
				PossibleAnswers: []int{1, 2},
				AnswersText:     []string{"a", "b"},
			},
			wantErr: true,
		},
		{
			name: "fail, empty text",
			fields: fields{
				Text:            "",
				AnswerType:      entity.AnswerTypeSegment,
				PossibleAnswers: []int{1, 2},
			},
			wantErr: true,
		},
		{
			name: "fail, empty possible answers",
			fields: fields{
				Text:            "Question text.",
				AnswerType:      entity.AnswerTypeSelect,
				PossibleAnswers: []int{},
			},
			wantErr: true,
		},
		{
			name: "fail, mismatched length",
			fields: fields{
				Text:            "Question text.",
				AnswerType:      entity.AnswerTypeSelect,
				PossibleAnswers: []int{1, 2},
				AnswersText:     []string{"a"},
			},
			wantErr: true,
		},
		{
			name: "fail, no empty answers text",
			fields: fields{
				Text:            "Question text.",
				AnswerType:      entity.AnswerTypeSegment,
				PossibleAnswers: []int{1, 2},
				AnswersText:     []string{"a", ""},
			},
			wantErr: true,
		},
		{
			name: "fail, wrong first letter",
			fields: fields{
				Text:            "question text.",
				AnswerType:      entity.AnswerTypeSegment,
				PossibleAnswers: []int{1, 3},
			},
			wantErr: true,
		},
		{
			name: "fail, wrong first letter",
			fields: fields{
				Text:            "1uestion text",
				AnswerType:      entity.AnswerTypeSegment,
				PossibleAnswers: []int{1, 3},
			},
			wantErr: true,
		},
		{
			name: "fail, wrong last letter",
			fields: fields{
				Text:            "question text",
				AnswerType:      entity.AnswerTypeSegment,
				PossibleAnswers: []int{1, 3},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := entity.Question{
				Text:            tt.fields.Text,
				AnswerType:      tt.fields.AnswerType,
				PossibleAnswers: tt.fields.PossibleAnswers,
				AnswersText:     tt.fields.AnswersText,
			}
			if err := q.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Question.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
