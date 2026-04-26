package layoutcatalog

import "testing"

func TestValidateValidHero(t *testing.T) {
	c := NewCatalog()
	if err := c.Load(); err != nil {
		t.Fatal(err)
	}
	md := ":::hero\neyebrow: 深度观察\ntitle: 真问题\n:::\n"
	r := c.Validate(md)
	if len(r.Errors) != 0 {
		t.Errorf("expected no errors, got %v", r.Errors)
	}
}

func TestValidateMissingRequiredField(t *testing.T) {
	c := NewCatalog()
	if err := c.Load(); err != nil {
		t.Fatal(err)
	}
	md := ":::hero\neyebrow: x\n:::\n"
	r := c.Validate(md)
	if len(r.Errors) == 0 {
		t.Fatalf("expected error for missing title")
	}
	if r.Errors[0].Module != "hero" || r.Errors[0].Field != "title" {
		t.Errorf("unexpected error: %+v", r.Errors[0])
	}
}

func TestValidateUnknownModuleWarns(t *testing.T) {
	c := NewCatalog()
	if err := c.Load(); err != nil {
		t.Fatal(err)
	}
	md := ":::futuristic-block\nfoo: bar\n:::\n"
	r := c.Validate(md)
	if len(r.Errors) != 0 {
		t.Errorf("unknown module must NOT error, got %v", r.Errors)
	}
	if len(r.Warnings) != 1 || r.Warnings[0].Module != "futuristic-block" {
		t.Errorf("expected one warning for futuristic-block, got %+v", r.Warnings)
	}
}
