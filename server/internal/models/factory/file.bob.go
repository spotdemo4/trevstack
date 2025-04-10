// Code generated by BobGen sql (devel). DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package factory

import (
	"context"
	"testing"

	"github.com/aarondl/opt/null"
	"github.com/aarondl/opt/omit"
	"github.com/jaswdr/faker/v2"
	models "github.com/spotdemo4/trevstack/server/internal/models"
	"github.com/stephenafamo/bob"
)

type FileMod interface {
	Apply(*FileTemplate)
}

type FileModFunc func(*FileTemplate)

func (f FileModFunc) Apply(n *FileTemplate) {
	f(n)
}

type FileModSlice []FileMod

func (mods FileModSlice) Apply(n *FileTemplate) {
	for _, f := range mods {
		f.Apply(n)
	}
}

// FileTemplate is an object representing the database table.
// all columns are optional and should be set by mods
type FileTemplate struct {
	ID     func() int64
	Name   func() string
	Data   func() []byte
	UserID func() int64

	r fileR
	f *Factory
}

type fileR struct {
	User                *fileRUserR
	ProfilePictureUsers []*fileRProfilePictureUsersR
}

type fileRUserR struct {
	o *UserTemplate
}
type fileRProfilePictureUsersR struct {
	number int
	o      *UserTemplate
}

// Apply mods to the FileTemplate
func (o *FileTemplate) Apply(mods ...FileMod) {
	for _, mod := range mods {
		mod.Apply(o)
	}
}

// toModel returns an *models.File
// this does nothing with the relationship templates
func (o FileTemplate) toModel() *models.File {
	m := &models.File{}

	if o.ID != nil {
		m.ID = o.ID()
	}
	if o.Name != nil {
		m.Name = o.Name()
	}
	if o.Data != nil {
		m.Data = o.Data()
	}
	if o.UserID != nil {
		m.UserID = o.UserID()
	}

	return m
}

// toModels returns an models.FileSlice
// this does nothing with the relationship templates
func (o FileTemplate) toModels(number int) models.FileSlice {
	m := make(models.FileSlice, number)

	for i := range m {
		m[i] = o.toModel()
	}

	return m
}

// setModelRels creates and sets the relationships on *models.File
// according to the relationships in the template. Nothing is inserted into the db
func (t FileTemplate) setModelRels(o *models.File) {
	if t.r.User != nil {
		rel := t.r.User.o.toModel()
		rel.R.Files = append(rel.R.Files, o)
		o.UserID = rel.ID
		o.R.User = rel
	}

	if t.r.ProfilePictureUsers != nil {
		rel := models.UserSlice{}
		for _, r := range t.r.ProfilePictureUsers {
			related := r.o.toModels(r.number)
			for _, rel := range related {
				rel.ProfilePictureID = null.From(o.ID)
				rel.R.ProfilePictureFile = o
			}
			rel = append(rel, related...)
		}
		o.R.ProfilePictureUsers = rel
	}
}

// BuildSetter returns an *models.FileSetter
// this does nothing with the relationship templates
func (o FileTemplate) BuildSetter() *models.FileSetter {
	m := &models.FileSetter{}

	if o.ID != nil {
		m.ID = omit.From(o.ID())
	}
	if o.Name != nil {
		m.Name = omit.From(o.Name())
	}
	if o.Data != nil {
		m.Data = omit.From(o.Data())
	}
	if o.UserID != nil {
		m.UserID = omit.From(o.UserID())
	}

	return m
}

// BuildManySetter returns an []*models.FileSetter
// this does nothing with the relationship templates
func (o FileTemplate) BuildManySetter(number int) []*models.FileSetter {
	m := make([]*models.FileSetter, number)

	for i := range m {
		m[i] = o.BuildSetter()
	}

	return m
}

// Build returns an *models.File
// Related objects are also created and placed in the .R field
// NOTE: Objects are not inserted into the database. Use FileTemplate.Create
func (o FileTemplate) Build() *models.File {
	m := o.toModel()
	o.setModelRels(m)

	return m
}

// BuildMany returns an models.FileSlice
// Related objects are also created and placed in the .R field
// NOTE: Objects are not inserted into the database. Use FileTemplate.CreateMany
func (o FileTemplate) BuildMany(number int) models.FileSlice {
	m := make(models.FileSlice, number)

	for i := range m {
		m[i] = o.Build()
	}

	return m
}

func ensureCreatableFile(m *models.FileSetter) {
	if m.Name.IsUnset() {
		m.Name = omit.From(random_string(nil))
	}
	if m.Data.IsUnset() {
		m.Data = omit.From(random___byte(nil))
	}
	if m.UserID.IsUnset() {
		m.UserID = omit.From(random_int64(nil))
	}
}

// insertOptRels creates and inserts any optional the relationships on *models.File
// according to the relationships in the template.
// any required relationship should have already exist on the model
func (o *FileTemplate) insertOptRels(ctx context.Context, exec bob.Executor, m *models.File) (context.Context, error) {
	var err error

	if o.r.ProfilePictureUsers != nil {
		for _, r := range o.r.ProfilePictureUsers {
			var rel1 models.UserSlice
			ctx, rel1, err = r.o.createMany(ctx, exec, r.number)
			if err != nil {
				return ctx, err
			}

			err = m.AttachProfilePictureUsers(ctx, exec, rel1...)
			if err != nil {
				return ctx, err
			}
		}
	}

	return ctx, err
}

// Create builds a file and inserts it into the database
// Relations objects are also inserted and placed in the .R field
func (o *FileTemplate) Create(ctx context.Context, exec bob.Executor) (*models.File, error) {
	_, m, err := o.create(ctx, exec)
	return m, err
}

// MustCreate builds a file and inserts it into the database
// Relations objects are also inserted and placed in the .R field
// panics if an error occurs
func (o *FileTemplate) MustCreate(ctx context.Context, exec bob.Executor) *models.File {
	_, m, err := o.create(ctx, exec)
	if err != nil {
		panic(err)
	}
	return m
}

// CreateOrFail builds a file and inserts it into the database
// Relations objects are also inserted and placed in the .R field
// It calls `tb.Fatal(err)` on the test/benchmark if an error occurs
func (o *FileTemplate) CreateOrFail(ctx context.Context, tb testing.TB, exec bob.Executor) *models.File {
	tb.Helper()
	_, m, err := o.create(ctx, exec)
	if err != nil {
		tb.Fatal(err)
		return nil
	}
	return m
}

// create builds a file and inserts it into the database
// Relations objects are also inserted and placed in the .R field
// this returns a context that includes the newly inserted model
func (o *FileTemplate) create(ctx context.Context, exec bob.Executor) (context.Context, *models.File, error) {
	var err error
	opt := o.BuildSetter()
	ensureCreatableFile(opt)

	var rel0 *models.User
	if o.r.User == nil {
		var ok bool
		rel0, ok = userCtx.Value(ctx)
		if !ok {
			FileMods.WithNewUser().Apply(o)
		}
	}
	if o.r.User != nil {
		ctx, rel0, err = o.r.User.o.create(ctx, exec)
		if err != nil {
			return ctx, nil, err
		}
	}
	opt.UserID = omit.From(rel0.ID)

	m, err := models.Files.Insert(opt).One(ctx, exec)
	if err != nil {
		return ctx, nil, err
	}
	ctx = fileCtx.WithValue(ctx, m)

	m.R.User = rel0

	ctx, err = o.insertOptRels(ctx, exec, m)
	return ctx, m, err
}

// CreateMany builds multiple files and inserts them into the database
// Relations objects are also inserted and placed in the .R field
func (o FileTemplate) CreateMany(ctx context.Context, exec bob.Executor, number int) (models.FileSlice, error) {
	_, m, err := o.createMany(ctx, exec, number)
	return m, err
}

// MustCreateMany builds multiple files and inserts them into the database
// Relations objects are also inserted and placed in the .R field
// panics if an error occurs
func (o FileTemplate) MustCreateMany(ctx context.Context, exec bob.Executor, number int) models.FileSlice {
	_, m, err := o.createMany(ctx, exec, number)
	if err != nil {
		panic(err)
	}
	return m
}

// CreateManyOrFail builds multiple files and inserts them into the database
// Relations objects are also inserted and placed in the .R field
// It calls `tb.Fatal(err)` on the test/benchmark if an error occurs
func (o FileTemplate) CreateManyOrFail(ctx context.Context, tb testing.TB, exec bob.Executor, number int) models.FileSlice {
	tb.Helper()
	_, m, err := o.createMany(ctx, exec, number)
	if err != nil {
		tb.Fatal(err)
		return nil
	}
	return m
}

// createMany builds multiple files and inserts them into the database
// Relations objects are also inserted and placed in the .R field
// this returns a context that includes the newly inserted models
func (o FileTemplate) createMany(ctx context.Context, exec bob.Executor, number int) (context.Context, models.FileSlice, error) {
	var err error
	m := make(models.FileSlice, number)

	for i := range m {
		ctx, m[i], err = o.create(ctx, exec)
		if err != nil {
			return ctx, nil, err
		}
	}

	return ctx, m, nil
}

// File has methods that act as mods for the FileTemplate
var FileMods fileMods

type fileMods struct{}

func (m fileMods) RandomizeAllColumns(f *faker.Faker) FileMod {
	return FileModSlice{
		FileMods.RandomID(f),
		FileMods.RandomName(f),
		FileMods.RandomData(f),
		FileMods.RandomUserID(f),
	}
}

// Set the model columns to this value
func (m fileMods) ID(val int64) FileMod {
	return FileModFunc(func(o *FileTemplate) {
		o.ID = func() int64 { return val }
	})
}

// Set the Column from the function
func (m fileMods) IDFunc(f func() int64) FileMod {
	return FileModFunc(func(o *FileTemplate) {
		o.ID = f
	})
}

// Clear any values for the column
func (m fileMods) UnsetID() FileMod {
	return FileModFunc(func(o *FileTemplate) {
		o.ID = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m fileMods) RandomID(f *faker.Faker) FileMod {
	return FileModFunc(func(o *FileTemplate) {
		o.ID = func() int64 {
			return random_int64(f)
		}
	})
}

// Set the model columns to this value
func (m fileMods) Name(val string) FileMod {
	return FileModFunc(func(o *FileTemplate) {
		o.Name = func() string { return val }
	})
}

// Set the Column from the function
func (m fileMods) NameFunc(f func() string) FileMod {
	return FileModFunc(func(o *FileTemplate) {
		o.Name = f
	})
}

// Clear any values for the column
func (m fileMods) UnsetName() FileMod {
	return FileModFunc(func(o *FileTemplate) {
		o.Name = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m fileMods) RandomName(f *faker.Faker) FileMod {
	return FileModFunc(func(o *FileTemplate) {
		o.Name = func() string {
			return random_string(f)
		}
	})
}

// Set the model columns to this value
func (m fileMods) Data(val []byte) FileMod {
	return FileModFunc(func(o *FileTemplate) {
		o.Data = func() []byte { return val }
	})
}

// Set the Column from the function
func (m fileMods) DataFunc(f func() []byte) FileMod {
	return FileModFunc(func(o *FileTemplate) {
		o.Data = f
	})
}

// Clear any values for the column
func (m fileMods) UnsetData() FileMod {
	return FileModFunc(func(o *FileTemplate) {
		o.Data = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m fileMods) RandomData(f *faker.Faker) FileMod {
	return FileModFunc(func(o *FileTemplate) {
		o.Data = func() []byte {
			return random___byte(f)
		}
	})
}

// Set the model columns to this value
func (m fileMods) UserID(val int64) FileMod {
	return FileModFunc(func(o *FileTemplate) {
		o.UserID = func() int64 { return val }
	})
}

// Set the Column from the function
func (m fileMods) UserIDFunc(f func() int64) FileMod {
	return FileModFunc(func(o *FileTemplate) {
		o.UserID = f
	})
}

// Clear any values for the column
func (m fileMods) UnsetUserID() FileMod {
	return FileModFunc(func(o *FileTemplate) {
		o.UserID = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m fileMods) RandomUserID(f *faker.Faker) FileMod {
	return FileModFunc(func(o *FileTemplate) {
		o.UserID = func() int64 {
			return random_int64(f)
		}
	})
}

func (m fileMods) WithUser(rel *UserTemplate) FileMod {
	return FileModFunc(func(o *FileTemplate) {
		o.r.User = &fileRUserR{
			o: rel,
		}
	})
}

func (m fileMods) WithNewUser(mods ...UserMod) FileMod {
	return FileModFunc(func(o *FileTemplate) {
		related := o.f.NewUser(mods...)

		m.WithUser(related).Apply(o)
	})
}

func (m fileMods) WithoutUser() FileMod {
	return FileModFunc(func(o *FileTemplate) {
		o.r.User = nil
	})
}

func (m fileMods) WithProfilePictureUsers(number int, related *UserTemplate) FileMod {
	return FileModFunc(func(o *FileTemplate) {
		o.r.ProfilePictureUsers = []*fileRProfilePictureUsersR{{
			number: number,
			o:      related,
		}}
	})
}

func (m fileMods) WithNewProfilePictureUsers(number int, mods ...UserMod) FileMod {
	return FileModFunc(func(o *FileTemplate) {
		related := o.f.NewUser(mods...)
		m.WithProfilePictureUsers(number, related).Apply(o)
	})
}

func (m fileMods) AddProfilePictureUsers(number int, related *UserTemplate) FileMod {
	return FileModFunc(func(o *FileTemplate) {
		o.r.ProfilePictureUsers = append(o.r.ProfilePictureUsers, &fileRProfilePictureUsersR{
			number: number,
			o:      related,
		})
	})
}

func (m fileMods) AddNewProfilePictureUsers(number int, mods ...UserMod) FileMod {
	return FileModFunc(func(o *FileTemplate) {
		related := o.f.NewUser(mods...)
		m.AddProfilePictureUsers(number, related).Apply(o)
	})
}

func (m fileMods) WithoutProfilePictureUsers() FileMod {
	return FileModFunc(func(o *FileTemplate) {
		o.r.ProfilePictureUsers = nil
	})
}
