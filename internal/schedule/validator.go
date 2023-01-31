package schedule

import (
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"studyum/pkg/datetime"
)

type Validator interface {
	AddGeneralLesson(dto AddGeneralLessonDTO) error

	AddLesson(dto AddLessonDTO) error
	UpdateLesson(lessonDTO UpdateLessonDTO) error
}

type schedule struct {
	validate *validator.Validate
}

func NewSchedule(validate *validator.Validate) Validator {
	return &schedule{validate: validate}
}

func (s *schedule) AddGeneralLesson(dto AddGeneralLessonDTO) error {
	startDuration, err := datetime.ParseDuration(dto.StartTime)
	if err != nil {
		return err
	}

	endDuration, err := datetime.ParseDuration(dto.EndTime)
	if err != nil {
		return err
	}

	if endDuration <= startDuration {
		return errors.Wrap(ValidationError, "start time is after end time")
	}

	return nil
}

func (s *schedule) AddLesson(dto AddLessonDTO) error {
	if !dto.StartDate.Before(dto.EndDate) {
		return errors.Wrap(ValidationError, "start date is after end date")
	}

	return nil
}

func (s *schedule) UpdateLesson(dto UpdateLessonDTO) error {
	if err := s.AddLesson(dto.AddLessonDTO); err != nil {
		return err
	}

	if dto.Id.IsZero() {
		return errors.Wrap(ValidationError, "not valid id")
	}

	return nil
}
