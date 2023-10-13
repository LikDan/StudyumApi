package controllers

import (
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/pkg/errors"
	"regexp"
	"studyum/internal/i18n/entities"
	"studyum/internal/i18n/repositories"
)

var (
	BadLangErr  = errors.New("Bad language was selected")
	BadGroupErr = errors.New("Bad group was selected")
)

type Controller interface {
	LoadDefaults(ctx context.Context, lang string) (entities.I18nWithHash, error)
	LoadByGroup(ctx context.Context, lang string, code string) (entities.I18nWithHash, error)
}

type controller struct {
	repository repositories.Repository

	i18nRegex      *regexp.Regexp
	codeGroupRegex *regexp.Regexp
}

func NewController(repository repositories.Repository) Controller {
	i18nRegex, err := regexp.Compile("^[a-z]{2}_[a-z]{2}$")
	if err != nil {
		return nil
	}

	codeGroupRegex, err := regexp.Compile("^[a-zA-Z_.]+$")
	if err != nil {
		return nil
	}

	return &controller{repository: repository, i18nRegex: i18nRegex, codeGroupRegex: codeGroupRegex}
}

func (c *controller) LoadDefaults(ctx context.Context, lang string) (entities.I18nWithHash, error) {
	if !c.i18nRegex.MatchString(lang) {
		return entities.I18nWithHash{}, BadLangErr
	}

	i18n, err := c.repository.GetByCode(ctx, lang, "defaults")
	if err != nil {
		return entities.I18nWithHash{}, err
	}

	return c.generateWithHash(ctx, i18n)
}

func (c *controller) LoadByGroup(ctx context.Context, lang string, group string) (entities.I18nWithHash, error) {
	if !c.i18nRegex.MatchString(lang) {
		return entities.I18nWithHash{}, BadLangErr
	}

	if !c.codeGroupRegex.MatchString(group) {
		return entities.I18nWithHash{}, BadGroupErr
	}

	i18n, err := c.repository.GetByCode(ctx, lang, group)
	if err != nil {
		return entities.I18nWithHash{}, err
	}

	return c.generateWithHash(ctx, i18n)
}

func (c *controller) generateWithHash(_ context.Context, i18n entities.I18n) (entities.I18nWithHash, error) {
	h := sha256.New()
	for k, v := range i18n {
		b := sha256.Sum256([]byte(fmt.Sprintf("%v:%v", k, v)))
		h.Write(b[:])
	}

	return entities.I18nWithHash{
		Hash:        fmt.Sprintf("%x", h.Sum(nil)),
		Translation: i18n,
	}, nil
}
