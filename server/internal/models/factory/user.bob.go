// Code generated by BobGen sql (devel). DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package factory

import (
	"context"
	"testing"

	"github.com/aarondl/opt/null"
	"github.com/aarondl/opt/omit"
	"github.com/aarondl/opt/omitnull"
	"github.com/jaswdr/faker/v2"
	models "github.com/spotdemo4/trevstack/server/internal/models"
	"github.com/stephenafamo/bob"
)

type UserMod interface {
	Apply(*UserTemplate)
}

type UserModFunc func(*UserTemplate)

func (f UserModFunc) Apply(n *UserTemplate) {
	f(n)
}

type UserModSlice []UserMod

func (mods UserModSlice) Apply(n *UserTemplate) {
	for _, f := range mods {
		f.Apply(n)
	}
}

// UserTemplate is an object representing the database table.
// all columns are optional and should be set by mods
type UserTemplate struct {
	ID               func() int64
	Username         func() string
	Password         func() string
	ProfilePictureID func() null.Val[int64]

	r userR
	f *Factory
}

type userR struct {
	Files              []*userRFilesR
	Items              []*userRItemsR
	ProfilePictureFile *userRProfilePictureFileR
}

type userRFilesR struct {
	number int
	o      *FileTemplate
}
type userRItemsR struct {
	number int
	o      *ItemTemplate
}
type userRProfilePictureFileR struct {
	o *FileTemplate
}

// Apply mods to the UserTemplate
func (o *UserTemplate) Apply(mods ...UserMod) {
	for _, mod := range mods {
		mod.Apply(o)
	}
}

// toModel returns an *models.User
// this does nothing with the relationship templates
func (o UserTemplate) toModel() *models.User {
	m := &models.User{}

	if o.ID != nil {
		m.ID = o.ID()
	}
	if o.Username != nil {
		m.Username = o.Username()
	}
	if o.Password != nil {
		m.Password = o.Password()
	}
	if o.ProfilePictureID != nil {
		m.ProfilePictureID = o.ProfilePictureID()
	}

	return m
}

// toModels returns an models.UserSlice
// this does nothing with the relationship templates
func (o UserTemplate) toModels(number int) models.UserSlice {
	m := make(models.UserSlice, number)

	for i := range m {
		m[i] = o.toModel()
	}

	return m
}

// setModelRels creates and sets the relationships on *models.User
// according to the relationships in the template. Nothing is inserted into the db
func (t UserTemplate) setModelRels(o *models.User) {
	if t.r.Files != nil {
		rel := models.FileSlice{}
		for _, r := range t.r.Files {
			related := r.o.toModels(r.number)
			for _, rel := range related {
				rel.UserID = o.ID
				rel.R.User = o
			}
			rel = append(rel, related...)
		}
		o.R.Files = rel
	}

	if t.r.Items != nil {
		rel := models.ItemSlice{}
		for _, r := range t.r.Items {
			related := r.o.toModels(r.number)
			for _, rel := range related {
				rel.UserID = o.ID
				rel.R.User = o
			}
			rel = append(rel, related...)
		}
		o.R.Items = rel
	}

	if t.r.ProfilePictureFile != nil {
		rel := t.r.ProfilePictureFile.o.toModel()
		rel.R.ProfilePictureUsers = append(rel.R.ProfilePictureUsers, o)
		o.ProfilePictureID = null.From(rel.ID)
		o.R.ProfilePictureFile = rel
	}
}

// BuildSetter returns an *models.UserSetter
// this does nothing with the relationship templates
func (o UserTemplate) BuildSetter() *models.UserSetter {
	m := &models.UserSetter{}

	if o.ID != nil {
		m.ID = omit.From(o.ID())
	}
	if o.Username != nil {
		m.Username = omit.From(o.Username())
	}
	if o.Password != nil {
		m.Password = omit.From(o.Password())
	}
	if o.ProfilePictureID != nil {
		m.ProfilePictureID = omitnull.FromNull(o.ProfilePictureID())
	}

	return m
}

// BuildManySetter returns an []*models.UserSetter
// this does nothing with the relationship templates
func (o UserTemplate) BuildManySetter(number int) []*models.UserSetter {
	m := make([]*models.UserSetter, number)

	for i := range m {
		m[i] = o.BuildSetter()
	}

	return m
}

// Build returns an *models.User
// Related objects are also created and placed in the .R field
// NOTE: Objects are not inserted into the database. Use UserTemplate.Create
func (o UserTemplate) Build() *models.User {
	m := o.toModel()
	o.setModelRels(m)

	return m
}

// BuildMany returns an models.UserSlice
// Related objects are also created and placed in the .R field
// NOTE: Objects are not inserted into the database. Use UserTemplate.CreateMany
func (o UserTemplate) BuildMany(number int) models.UserSlice {
	m := make(models.UserSlice, number)

	for i := range m {
		m[i] = o.Build()
	}

	return m
}

func ensureCreatableUser(m *models.UserSetter) {
	if m.Username.IsUnset() {
		m.Username = omit.From(random_string(nil))
	}
	if m.Password.IsUnset() {
		m.Password = omit.From(random_string(nil))
	}
}

// insertOptRels creates and inserts any optional the relationships on *models.User
// according to the relationships in the template.
// any required relationship should have already exist on the model
func (o *UserTemplate) insertOptRels(ctx context.Context, exec bob.Executor, m *models.User) (context.Context, error) {
	var err error

	if o.r.Files != nil {
		for _, r := range o.r.Files {
			var rel0 models.FileSlice
			ctx, rel0, err = r.o.createMany(ctx, exec, r.number)
			if err != nil {
				return ctx, err
			}

			err = m.AttachFiles(ctx, exec, rel0...)
			if err != nil {
				return ctx, err
			}
		}
	}

	if o.r.Items != nil {
		for _, r := range o.r.Items {
			var rel1 models.ItemSlice
			ctx, rel1, err = r.o.createMany(ctx, exec, r.number)
			if err != nil {
				return ctx, err
			}

			err = m.AttachItems(ctx, exec, rel1...)
			if err != nil {
				return ctx, err
			}
		}
	}

	if o.r.ProfilePictureFile != nil {
		var rel2 *models.File
		ctx, rel2, err = o.r.ProfilePictureFile.o.create(ctx, exec)
		if err != nil {
			return ctx, err
		}
		err = m.AttachProfilePictureFile(ctx, exec, rel2)
		if err != nil {
			return ctx, err
		}
	}

	return ctx, err
}

// Create builds a user and inserts it into the database
// Relations objects are also inserted and placed in the .R field
func (o *UserTemplate) Create(ctx context.Context, exec bob.Executor) (*models.User, error) {
	_, m, err := o.create(ctx, exec)
	return m, err
}

// MustCreate builds a user and inserts it into the database
// Relations objects are also inserted and placed in the .R field
// panics if an error occurs
func (o *UserTemplate) MustCreate(ctx context.Context, exec bob.Executor) *models.User {
	_, m, err := o.create(ctx, exec)
	if err != nil {
		panic(err)
	}
	return m
}

// CreateOrFail builds a user and inserts it into the database
// Relations objects are also inserted and placed in the .R field
// It calls `tb.Fatal(err)` on the test/benchmark if an error occurs
func (o *UserTemplate) CreateOrFail(ctx context.Context, tb testing.TB, exec bob.Executor) *models.User {
	tb.Helper()
	_, m, err := o.create(ctx, exec)
	if err != nil {
		tb.Fatal(err)
		return nil
	}
	return m
}

// create builds a user and inserts it into the database
// Relations objects are also inserted and placed in the .R field
// this returns a context that includes the newly inserted model
func (o *UserTemplate) create(ctx context.Context, exec bob.Executor) (context.Context, *models.User, error) {
	var err error
	opt := o.BuildSetter()
	ensureCreatableUser(opt)

	m, err := models.Users.Insert(opt).One(ctx, exec)
	if err != nil {
		return ctx, nil, err
	}
	ctx = userCtx.WithValue(ctx, m)

	ctx, err = o.insertOptRels(ctx, exec, m)
	return ctx, m, err
}

// CreateMany builds multiple users and inserts them into the database
// Relations objects are also inserted and placed in the .R field
func (o UserTemplate) CreateMany(ctx context.Context, exec bob.Executor, number int) (models.UserSlice, error) {
	_, m, err := o.createMany(ctx, exec, number)
	return m, err
}

// MustCreateMany builds multiple users and inserts them into the database
// Relations objects are also inserted and placed in the .R field
// panics if an error occurs
func (o UserTemplate) MustCreateMany(ctx context.Context, exec bob.Executor, number int) models.UserSlice {
	_, m, err := o.createMany(ctx, exec, number)
	if err != nil {
		panic(err)
	}
	return m
}

// CreateManyOrFail builds multiple users and inserts them into the database
// Relations objects are also inserted and placed in the .R field
// It calls `tb.Fatal(err)` on the test/benchmark if an error occurs
func (o UserTemplate) CreateManyOrFail(ctx context.Context, tb testing.TB, exec bob.Executor, number int) models.UserSlice {
	tb.Helper()
	_, m, err := o.createMany(ctx, exec, number)
	if err != nil {
		tb.Fatal(err)
		return nil
	}
	return m
}

// createMany builds multiple users and inserts them into the database
// Relations objects are also inserted and placed in the .R field
// this returns a context that includes the newly inserted models
func (o UserTemplate) createMany(ctx context.Context, exec bob.Executor, number int) (context.Context, models.UserSlice, error) {
	var err error
	m := make(models.UserSlice, number)

	for i := range m {
		ctx, m[i], err = o.create(ctx, exec)
		if err != nil {
			return ctx, nil, err
		}
	}

	return ctx, m, nil
}

// User has methods that act as mods for the UserTemplate
var UserMods userMods

type userMods struct{}

func (m userMods) RandomizeAllColumns(f *faker.Faker) UserMod {
	return UserModSlice{
		UserMods.RandomID(f),
		UserMods.RandomUsername(f),
		UserMods.RandomPassword(f),
		UserMods.RandomProfilePictureID(f),
	}
}

// Set the model columns to this value
func (m userMods) ID(val int64) UserMod {
	return UserModFunc(func(o *UserTemplate) {
		o.ID = func() int64 { return val }
	})
}

// Set the Column from the function
func (m userMods) IDFunc(f func() int64) UserMod {
	return UserModFunc(func(o *UserTemplate) {
		o.ID = f
	})
}

// Clear any values for the column
func (m userMods) UnsetID() UserMod {
	return UserModFunc(func(o *UserTemplate) {
		o.ID = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m userMods) RandomID(f *faker.Faker) UserMod {
	return UserModFunc(func(o *UserTemplate) {
		o.ID = func() int64 {
			return random_int64(f)
		}
	})
}

// Set the model columns to this value
func (m userMods) Username(val string) UserMod {
	return UserModFunc(func(o *UserTemplate) {
		o.Username = func() string { return val }
	})
}

// Set the Column from the function
func (m userMods) UsernameFunc(f func() string) UserMod {
	return UserModFunc(func(o *UserTemplate) {
		o.Username = f
	})
}

// Clear any values for the column
func (m userMods) UnsetUsername() UserMod {
	return UserModFunc(func(o *UserTemplate) {
		o.Username = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m userMods) RandomUsername(f *faker.Faker) UserMod {
	return UserModFunc(func(o *UserTemplate) {
		o.Username = func() string {
			return random_string(f)
		}
	})
}

// Set the model columns to this value
func (m userMods) Password(val string) UserMod {
	return UserModFunc(func(o *UserTemplate) {
		o.Password = func() string { return val }
	})
}

// Set the Column from the function
func (m userMods) PasswordFunc(f func() string) UserMod {
	return UserModFunc(func(o *UserTemplate) {
		o.Password = f
	})
}

// Clear any values for the column
func (m userMods) UnsetPassword() UserMod {
	return UserModFunc(func(o *UserTemplate) {
		o.Password = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m userMods) RandomPassword(f *faker.Faker) UserMod {
	return UserModFunc(func(o *UserTemplate) {
		o.Password = func() string {
			return random_string(f)
		}
	})
}

// Set the model columns to this value
func (m userMods) ProfilePictureID(val null.Val[int64]) UserMod {
	return UserModFunc(func(o *UserTemplate) {
		o.ProfilePictureID = func() null.Val[int64] { return val }
	})
}

// Set the Column from the function
func (m userMods) ProfilePictureIDFunc(f func() null.Val[int64]) UserMod {
	return UserModFunc(func(o *UserTemplate) {
		o.ProfilePictureID = f
	})
}

// Clear any values for the column
func (m userMods) UnsetProfilePictureID() UserMod {
	return UserModFunc(func(o *UserTemplate) {
		o.ProfilePictureID = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m userMods) RandomProfilePictureID(f *faker.Faker) UserMod {
	return UserModFunc(func(o *UserTemplate) {
		o.ProfilePictureID = func() null.Val[int64] {
			if f == nil {
				f = &defaultFaker
			}

			if f.Bool() {
				return null.FromPtr[int64](nil)
			}

			return null.From(random_int64(f))
		}
	})
}

func (m userMods) WithProfilePictureFile(rel *FileTemplate) UserMod {
	return UserModFunc(func(o *UserTemplate) {
		o.r.ProfilePictureFile = &userRProfilePictureFileR{
			o: rel,
		}
	})
}

func (m userMods) WithNewProfilePictureFile(mods ...FileMod) UserMod {
	return UserModFunc(func(o *UserTemplate) {
		related := o.f.NewFile(mods...)

		m.WithProfilePictureFile(related).Apply(o)
	})
}

func (m userMods) WithoutProfilePictureFile() UserMod {
	return UserModFunc(func(o *UserTemplate) {
		o.r.ProfilePictureFile = nil
	})
}

func (m userMods) WithFiles(number int, related *FileTemplate) UserMod {
	return UserModFunc(func(o *UserTemplate) {
		o.r.Files = []*userRFilesR{{
			number: number,
			o:      related,
		}}
	})
}

func (m userMods) WithNewFiles(number int, mods ...FileMod) UserMod {
	return UserModFunc(func(o *UserTemplate) {
		related := o.f.NewFile(mods...)
		m.WithFiles(number, related).Apply(o)
	})
}

func (m userMods) AddFiles(number int, related *FileTemplate) UserMod {
	return UserModFunc(func(o *UserTemplate) {
		o.r.Files = append(o.r.Files, &userRFilesR{
			number: number,
			o:      related,
		})
	})
}

func (m userMods) AddNewFiles(number int, mods ...FileMod) UserMod {
	return UserModFunc(func(o *UserTemplate) {
		related := o.f.NewFile(mods...)
		m.AddFiles(number, related).Apply(o)
	})
}

func (m userMods) WithoutFiles() UserMod {
	return UserModFunc(func(o *UserTemplate) {
		o.r.Files = nil
	})
}

func (m userMods) WithItems(number int, related *ItemTemplate) UserMod {
	return UserModFunc(func(o *UserTemplate) {
		o.r.Items = []*userRItemsR{{
			number: number,
			o:      related,
		}}
	})
}

func (m userMods) WithNewItems(number int, mods ...ItemMod) UserMod {
	return UserModFunc(func(o *UserTemplate) {
		related := o.f.NewItem(mods...)
		m.WithItems(number, related).Apply(o)
	})
}

func (m userMods) AddItems(number int, related *ItemTemplate) UserMod {
	return UserModFunc(func(o *UserTemplate) {
		o.r.Items = append(o.r.Items, &userRItemsR{
			number: number,
			o:      related,
		})
	})
}

func (m userMods) AddNewItems(number int, mods ...ItemMod) UserMod {
	return UserModFunc(func(o *UserTemplate) {
		related := o.f.NewItem(mods...)
		m.AddItems(number, related).Apply(o)
	})
}

func (m userMods) WithoutItems() UserMod {
	return UserModFunc(func(o *UserTemplate) {
		o.r.Items = nil
	})
}
