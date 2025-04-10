// Code generated by BobGen sql (devel). DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"

	"github.com/aarondl/opt/null"
	"github.com/aarondl/opt/omit"
	"github.com/aarondl/opt/omitnull"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/sqlite"
	"github.com/stephenafamo/bob/dialect/sqlite/dialect"
	"github.com/stephenafamo/bob/dialect/sqlite/dm"
	"github.com/stephenafamo/bob/dialect/sqlite/sm"
	"github.com/stephenafamo/bob/dialect/sqlite/um"
	"github.com/stephenafamo/bob/expr"
	"github.com/stephenafamo/bob/mods"
	"github.com/stephenafamo/bob/orm"
)

// User is an object representing the database table.
type User struct {
	ID               int64           `db:"id,pk" `
	Username         string          `db:"username" `
	Password         string          `db:"password" `
	ProfilePictureID null.Val[int64] `db:"profile_picture_id" `

	R userR `db:"-" `
}

// UserSlice is an alias for a slice of pointers to User.
// This should almost always be used instead of []*User.
type UserSlice []*User

// Users contains methods to work with the user table
var Users = sqlite.NewTablex[*User, UserSlice, *UserSetter]("", "user")

// UsersQuery is a query on the user table
type UsersQuery = *sqlite.ViewQuery[*User, UserSlice]

// userR is where relationships are stored.
type userR struct {
	Files              FileSlice // fk_file_0
	Items              ItemSlice // fk_item_0
	ProfilePictureFile *File     // fk_user_0
}

type userColumnNames struct {
	ID               string
	Username         string
	Password         string
	ProfilePictureID string
}

var UserColumns = buildUserColumns("user")

type userColumns struct {
	tableAlias       string
	ID               sqlite.Expression
	Username         sqlite.Expression
	Password         sqlite.Expression
	ProfilePictureID sqlite.Expression
}

func (c userColumns) Alias() string {
	return c.tableAlias
}

func (userColumns) AliasedAs(alias string) userColumns {
	return buildUserColumns(alias)
}

func buildUserColumns(alias string) userColumns {
	return userColumns{
		tableAlias:       alias,
		ID:               sqlite.Quote(alias, "id"),
		Username:         sqlite.Quote(alias, "username"),
		Password:         sqlite.Quote(alias, "password"),
		ProfilePictureID: sqlite.Quote(alias, "profile_picture_id"),
	}
}

type userWhere[Q sqlite.Filterable] struct {
	ID               sqlite.WhereMod[Q, int64]
	Username         sqlite.WhereMod[Q, string]
	Password         sqlite.WhereMod[Q, string]
	ProfilePictureID sqlite.WhereNullMod[Q, int64]
}

func (userWhere[Q]) AliasedAs(alias string) userWhere[Q] {
	return buildUserWhere[Q](buildUserColumns(alias))
}

func buildUserWhere[Q sqlite.Filterable](cols userColumns) userWhere[Q] {
	return userWhere[Q]{
		ID:               sqlite.Where[Q, int64](cols.ID),
		Username:         sqlite.Where[Q, string](cols.Username),
		Password:         sqlite.Where[Q, string](cols.Password),
		ProfilePictureID: sqlite.WhereNull[Q, int64](cols.ProfilePictureID),
	}
}

var UserErrors = &userErrors{
	ErrUniquePkMainUser: &UniqueConstraintError{s: "pk_main_user"},
}

type userErrors struct {
	ErrUniquePkMainUser *UniqueConstraintError
}

// UserSetter is used for insert/upsert/update operations
// All values are optional, and do not have to be set
// Generated columns are not included
type UserSetter struct {
	ID               omit.Val[int64]     `db:"id,pk" `
	Username         omit.Val[string]    `db:"username" `
	Password         omit.Val[string]    `db:"password" `
	ProfilePictureID omitnull.Val[int64] `db:"profile_picture_id" `
}

func (s UserSetter) SetColumns() []string {
	vals := make([]string, 0, 4)
	if !s.ID.IsUnset() {
		vals = append(vals, "id")
	}

	if !s.Username.IsUnset() {
		vals = append(vals, "username")
	}

	if !s.Password.IsUnset() {
		vals = append(vals, "password")
	}

	if !s.ProfilePictureID.IsUnset() {
		vals = append(vals, "profile_picture_id")
	}

	return vals
}

func (s UserSetter) Overwrite(t *User) {
	if !s.ID.IsUnset() {
		t.ID, _ = s.ID.Get()
	}
	if !s.Username.IsUnset() {
		t.Username, _ = s.Username.Get()
	}
	if !s.Password.IsUnset() {
		t.Password, _ = s.Password.Get()
	}
	if !s.ProfilePictureID.IsUnset() {
		t.ProfilePictureID, _ = s.ProfilePictureID.GetNull()
	}
}

func (s *UserSetter) Apply(q *dialect.InsertQuery) {
	q.AppendHooks(func(ctx context.Context, exec bob.Executor) (context.Context, error) {
		return Users.BeforeInsertHooks.RunHooks(ctx, exec, s)
	})

	if len(q.Table.Columns) == 0 {
		q.Table.Columns = s.SetColumns()
	}

	q.AppendValues(bob.ExpressionFunc(func(ctx context.Context, w io.Writer, d bob.Dialect, start int) ([]any, error) {
		vals := make([]bob.Expression, 0, 4)
		if !s.ID.IsUnset() {
			vals = append(vals, sqlite.Arg(s.ID))
		}

		if !s.Username.IsUnset() {
			vals = append(vals, sqlite.Arg(s.Username))
		}

		if !s.Password.IsUnset() {
			vals = append(vals, sqlite.Arg(s.Password))
		}

		if !s.ProfilePictureID.IsUnset() {
			vals = append(vals, sqlite.Arg(s.ProfilePictureID))
		}

		return bob.ExpressSlice(ctx, w, d, start, vals, "", ", ", "")
	}))
}

func (s UserSetter) UpdateMod() bob.Mod[*dialect.UpdateQuery] {
	return um.Set(s.Expressions()...)
}

func (s UserSetter) Expressions(prefix ...string) []bob.Expression {
	exprs := make([]bob.Expression, 0, 4)

	if !s.ID.IsUnset() {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			sqlite.Quote(append(prefix, "id")...),
			sqlite.Arg(s.ID),
		}})
	}

	if !s.Username.IsUnset() {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			sqlite.Quote(append(prefix, "username")...),
			sqlite.Arg(s.Username),
		}})
	}

	if !s.Password.IsUnset() {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			sqlite.Quote(append(prefix, "password")...),
			sqlite.Arg(s.Password),
		}})
	}

	if !s.ProfilePictureID.IsUnset() {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			sqlite.Quote(append(prefix, "profile_picture_id")...),
			sqlite.Arg(s.ProfilePictureID),
		}})
	}

	return exprs
}

// FindUser retrieves a single record by primary key
// If cols is empty Find will return all columns.
func FindUser(ctx context.Context, exec bob.Executor, IDPK int64, cols ...string) (*User, error) {
	if len(cols) == 0 {
		return Users.Query(
			SelectWhere.Users.ID.EQ(IDPK),
		).One(ctx, exec)
	}

	return Users.Query(
		SelectWhere.Users.ID.EQ(IDPK),
		sm.Columns(Users.Columns().Only(cols...)),
	).One(ctx, exec)
}

// UserExists checks the presence of a single record by primary key
func UserExists(ctx context.Context, exec bob.Executor, IDPK int64) (bool, error) {
	return Users.Query(
		SelectWhere.Users.ID.EQ(IDPK),
	).Exists(ctx, exec)
}

// AfterQueryHook is called after User is retrieved from the database
func (o *User) AfterQueryHook(ctx context.Context, exec bob.Executor, queryType bob.QueryType) error {
	var err error

	switch queryType {
	case bob.QueryTypeSelect:
		ctx, err = Users.AfterSelectHooks.RunHooks(ctx, exec, UserSlice{o})
	case bob.QueryTypeInsert:
		ctx, err = Users.AfterInsertHooks.RunHooks(ctx, exec, UserSlice{o})
	case bob.QueryTypeUpdate:
		ctx, err = Users.AfterUpdateHooks.RunHooks(ctx, exec, UserSlice{o})
	case bob.QueryTypeDelete:
		ctx, err = Users.AfterDeleteHooks.RunHooks(ctx, exec, UserSlice{o})
	}

	return err
}

// PrimaryKeyVals returns the primary key values of the User
func (o *User) PrimaryKeyVals() bob.Expression {
	return sqlite.Arg(o.ID)
}

func (o *User) pkEQ() dialect.Expression {
	return sqlite.Quote("user", "id").EQ(bob.ExpressionFunc(func(ctx context.Context, w io.Writer, d bob.Dialect, start int) ([]any, error) {
		return o.PrimaryKeyVals().WriteSQL(ctx, w, d, start)
	}))
}

// Update uses an executor to update the User
func (o *User) Update(ctx context.Context, exec bob.Executor, s *UserSetter) error {
	v, err := Users.Update(s.UpdateMod(), um.Where(o.pkEQ())).One(ctx, exec)
	if err != nil {
		return err
	}

	o.R = v.R
	*o = *v

	return nil
}

// Delete deletes a single User record with an executor
func (o *User) Delete(ctx context.Context, exec bob.Executor) error {
	_, err := Users.Delete(dm.Where(o.pkEQ())).Exec(ctx, exec)
	return err
}

// Reload refreshes the User using the executor
func (o *User) Reload(ctx context.Context, exec bob.Executor) error {
	o2, err := Users.Query(
		SelectWhere.Users.ID.EQ(o.ID),
	).One(ctx, exec)
	if err != nil {
		return err
	}
	o2.R = o.R
	*o = *o2

	return nil
}

// AfterQueryHook is called after UserSlice is retrieved from the database
func (o UserSlice) AfterQueryHook(ctx context.Context, exec bob.Executor, queryType bob.QueryType) error {
	var err error

	switch queryType {
	case bob.QueryTypeSelect:
		ctx, err = Users.AfterSelectHooks.RunHooks(ctx, exec, o)
	case bob.QueryTypeInsert:
		ctx, err = Users.AfterInsertHooks.RunHooks(ctx, exec, o)
	case bob.QueryTypeUpdate:
		ctx, err = Users.AfterUpdateHooks.RunHooks(ctx, exec, o)
	case bob.QueryTypeDelete:
		ctx, err = Users.AfterDeleteHooks.RunHooks(ctx, exec, o)
	}

	return err
}

func (o UserSlice) pkIN() dialect.Expression {
	if len(o) == 0 {
		return sqlite.Raw("NULL")
	}

	return sqlite.Quote("user", "id").In(bob.ExpressionFunc(func(ctx context.Context, w io.Writer, d bob.Dialect, start int) ([]any, error) {
		pkPairs := make([]bob.Expression, len(o))
		for i, row := range o {
			pkPairs[i] = row.PrimaryKeyVals()
		}
		return bob.ExpressSlice(ctx, w, d, start, pkPairs, "", ", ", "")
	}))
}

// copyMatchingRows finds models in the given slice that have the same primary key
// then it first copies the existing relationships from the old model to the new model
// and then replaces the old model in the slice with the new model
func (o UserSlice) copyMatchingRows(from ...*User) {
	for i, old := range o {
		for _, new := range from {
			if new.ID != old.ID {
				continue
			}
			new.R = old.R
			o[i] = new
			break
		}
	}
}

// UpdateMod modifies an update query with "WHERE primary_key IN (o...)"
func (o UserSlice) UpdateMod() bob.Mod[*dialect.UpdateQuery] {
	return bob.ModFunc[*dialect.UpdateQuery](func(q *dialect.UpdateQuery) {
		q.AppendHooks(func(ctx context.Context, exec bob.Executor) (context.Context, error) {
			return Users.BeforeUpdateHooks.RunHooks(ctx, exec, o)
		})

		q.AppendLoader(bob.LoaderFunc(func(ctx context.Context, exec bob.Executor, retrieved any) error {
			var err error
			switch retrieved := retrieved.(type) {
			case *User:
				o.copyMatchingRows(retrieved)
			case []*User:
				o.copyMatchingRows(retrieved...)
			case UserSlice:
				o.copyMatchingRows(retrieved...)
			default:
				// If the retrieved value is not a User or a slice of User
				// then run the AfterUpdateHooks on the slice
				_, err = Users.AfterUpdateHooks.RunHooks(ctx, exec, o)
			}

			return err
		}))

		q.AppendWhere(o.pkIN())
	})
}

// DeleteMod modifies an delete query with "WHERE primary_key IN (o...)"
func (o UserSlice) DeleteMod() bob.Mod[*dialect.DeleteQuery] {
	return bob.ModFunc[*dialect.DeleteQuery](func(q *dialect.DeleteQuery) {
		q.AppendHooks(func(ctx context.Context, exec bob.Executor) (context.Context, error) {
			return Users.BeforeDeleteHooks.RunHooks(ctx, exec, o)
		})

		q.AppendLoader(bob.LoaderFunc(func(ctx context.Context, exec bob.Executor, retrieved any) error {
			var err error
			switch retrieved := retrieved.(type) {
			case *User:
				o.copyMatchingRows(retrieved)
			case []*User:
				o.copyMatchingRows(retrieved...)
			case UserSlice:
				o.copyMatchingRows(retrieved...)
			default:
				// If the retrieved value is not a User or a slice of User
				// then run the AfterDeleteHooks on the slice
				_, err = Users.AfterDeleteHooks.RunHooks(ctx, exec, o)
			}

			return err
		}))

		q.AppendWhere(o.pkIN())
	})
}

func (o UserSlice) UpdateAll(ctx context.Context, exec bob.Executor, vals UserSetter) error {
	if len(o) == 0 {
		return nil
	}

	_, err := Users.Update(vals.UpdateMod(), o.UpdateMod()).All(ctx, exec)
	return err
}

func (o UserSlice) DeleteAll(ctx context.Context, exec bob.Executor) error {
	if len(o) == 0 {
		return nil
	}

	_, err := Users.Delete(o.DeleteMod()).Exec(ctx, exec)
	return err
}

func (o UserSlice) ReloadAll(ctx context.Context, exec bob.Executor) error {
	if len(o) == 0 {
		return nil
	}

	o2, err := Users.Query(sm.Where(o.pkIN())).All(ctx, exec)
	if err != nil {
		return err
	}

	o.copyMatchingRows(o2...)

	return nil
}

type userJoins[Q dialect.Joinable] struct {
	typ                string
	Files              func(context.Context) modAs[Q, fileColumns]
	Items              func(context.Context) modAs[Q, itemColumns]
	ProfilePictureFile func(context.Context) modAs[Q, fileColumns]
}

func (j userJoins[Q]) aliasedAs(alias string) userJoins[Q] {
	return buildUserJoins[Q](buildUserColumns(alias), j.typ)
}

func buildUserJoins[Q dialect.Joinable](cols userColumns, typ string) userJoins[Q] {
	return userJoins[Q]{
		typ:                typ,
		Files:              usersJoinFiles[Q](cols, typ),
		Items:              usersJoinItems[Q](cols, typ),
		ProfilePictureFile: usersJoinProfilePictureFile[Q](cols, typ),
	}
}

func usersJoinFiles[Q dialect.Joinable](from userColumns, typ string) func(context.Context) modAs[Q, fileColumns] {
	return func(ctx context.Context) modAs[Q, fileColumns] {
		return modAs[Q, fileColumns]{
			c: FileColumns,
			f: func(to fileColumns) bob.Mod[Q] {
				mods := make(mods.QueryMods[Q], 0, 1)

				{
					mods = append(mods, dialect.Join[Q](typ, Files.Name().As(to.Alias())).On(
						to.UserID.EQ(from.ID),
					))
				}

				return mods
			},
		}
	}
}

func usersJoinItems[Q dialect.Joinable](from userColumns, typ string) func(context.Context) modAs[Q, itemColumns] {
	return func(ctx context.Context) modAs[Q, itemColumns] {
		return modAs[Q, itemColumns]{
			c: ItemColumns,
			f: func(to itemColumns) bob.Mod[Q] {
				mods := make(mods.QueryMods[Q], 0, 1)

				{
					mods = append(mods, dialect.Join[Q](typ, Items.Name().As(to.Alias())).On(
						to.UserID.EQ(from.ID),
					))
				}

				return mods
			},
		}
	}
}

func usersJoinProfilePictureFile[Q dialect.Joinable](from userColumns, typ string) func(context.Context) modAs[Q, fileColumns] {
	return func(ctx context.Context) modAs[Q, fileColumns] {
		return modAs[Q, fileColumns]{
			c: FileColumns,
			f: func(to fileColumns) bob.Mod[Q] {
				mods := make(mods.QueryMods[Q], 0, 1)

				{
					mods = append(mods, dialect.Join[Q](typ, Files.Name().As(to.Alias())).On(
						to.ID.EQ(from.ProfilePictureID),
					))
				}

				return mods
			},
		}
	}
}

// Files starts a query for related objects on file
func (o *User) Files(mods ...bob.Mod[*dialect.SelectQuery]) FilesQuery {
	return Files.Query(append(mods,
		sm.Where(FileColumns.UserID.EQ(sqlite.Arg(o.ID))),
	)...)
}

func (os UserSlice) Files(mods ...bob.Mod[*dialect.SelectQuery]) FilesQuery {
	PKArgs := make([]bob.Expression, len(os))
	for i, o := range os {
		PKArgs[i] = sqlite.ArgGroup(o.ID)
	}

	return Files.Query(append(mods,
		sm.Where(sqlite.Group(FileColumns.UserID).In(PKArgs...)),
	)...)
}

// Items starts a query for related objects on item
func (o *User) Items(mods ...bob.Mod[*dialect.SelectQuery]) ItemsQuery {
	return Items.Query(append(mods,
		sm.Where(ItemColumns.UserID.EQ(sqlite.Arg(o.ID))),
	)...)
}

func (os UserSlice) Items(mods ...bob.Mod[*dialect.SelectQuery]) ItemsQuery {
	PKArgs := make([]bob.Expression, len(os))
	for i, o := range os {
		PKArgs[i] = sqlite.ArgGroup(o.ID)
	}

	return Items.Query(append(mods,
		sm.Where(sqlite.Group(ItemColumns.UserID).In(PKArgs...)),
	)...)
}

// ProfilePictureFile starts a query for related objects on file
func (o *User) ProfilePictureFile(mods ...bob.Mod[*dialect.SelectQuery]) FilesQuery {
	return Files.Query(append(mods,
		sm.Where(FileColumns.ID.EQ(sqlite.Arg(o.ProfilePictureID))),
	)...)
}

func (os UserSlice) ProfilePictureFile(mods ...bob.Mod[*dialect.SelectQuery]) FilesQuery {
	PKArgs := make([]bob.Expression, len(os))
	for i, o := range os {
		PKArgs[i] = sqlite.ArgGroup(o.ProfilePictureID)
	}

	return Files.Query(append(mods,
		sm.Where(sqlite.Group(FileColumns.ID).In(PKArgs...)),
	)...)
}

func (o *User) Preload(name string, retrieved any) error {
	if o == nil {
		return nil
	}

	switch name {
	case "Files":
		rels, ok := retrieved.(FileSlice)
		if !ok {
			return fmt.Errorf("user cannot load %T as %q", retrieved, name)
		}

		o.R.Files = rels

		for _, rel := range rels {
			if rel != nil {
				rel.R.User = o
			}
		}
		return nil
	case "Items":
		rels, ok := retrieved.(ItemSlice)
		if !ok {
			return fmt.Errorf("user cannot load %T as %q", retrieved, name)
		}

		o.R.Items = rels

		for _, rel := range rels {
			if rel != nil {
				rel.R.User = o
			}
		}
		return nil
	case "ProfilePictureFile":
		rel, ok := retrieved.(*File)
		if !ok {
			return fmt.Errorf("user cannot load %T as %q", retrieved, name)
		}

		o.R.ProfilePictureFile = rel

		if rel != nil {
			rel.R.ProfilePictureUsers = UserSlice{o}
		}
		return nil
	default:
		return fmt.Errorf("user has no relationship %q", name)
	}
}

func ThenLoadUserFiles(queryMods ...bob.Mod[*dialect.SelectQuery]) sqlite.Loader {
	return sqlite.Loader(func(ctx context.Context, exec bob.Executor, retrieved any) error {
		loader, isLoader := retrieved.(interface {
			LoadUserFiles(context.Context, bob.Executor, ...bob.Mod[*dialect.SelectQuery]) error
		})
		if !isLoader {
			return fmt.Errorf("object %T cannot load UserFiles", retrieved)
		}

		err := loader.LoadUserFiles(ctx, exec, queryMods...)

		// Don't cause an issue due to missing relationships
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}

		return err
	})
}

// LoadUserFiles loads the user's Files into the .R struct
func (o *User) LoadUserFiles(ctx context.Context, exec bob.Executor, mods ...bob.Mod[*dialect.SelectQuery]) error {
	if o == nil {
		return nil
	}

	// Reset the relationship
	o.R.Files = nil

	related, err := o.Files(mods...).All(ctx, exec)
	if err != nil {
		return err
	}

	for _, rel := range related {
		rel.R.User = o
	}

	o.R.Files = related
	return nil
}

// LoadUserFiles loads the user's Files into the .R struct
func (os UserSlice) LoadUserFiles(ctx context.Context, exec bob.Executor, mods ...bob.Mod[*dialect.SelectQuery]) error {
	if len(os) == 0 {
		return nil
	}

	files, err := os.Files(mods...).All(ctx, exec)
	if err != nil {
		return err
	}

	for _, o := range os {
		o.R.Files = nil
	}

	for _, o := range os {
		for _, rel := range files {
			if o.ID != rel.UserID {
				continue
			}

			rel.R.User = o

			o.R.Files = append(o.R.Files, rel)
		}
	}

	return nil
}

func ThenLoadUserItems(queryMods ...bob.Mod[*dialect.SelectQuery]) sqlite.Loader {
	return sqlite.Loader(func(ctx context.Context, exec bob.Executor, retrieved any) error {
		loader, isLoader := retrieved.(interface {
			LoadUserItems(context.Context, bob.Executor, ...bob.Mod[*dialect.SelectQuery]) error
		})
		if !isLoader {
			return fmt.Errorf("object %T cannot load UserItems", retrieved)
		}

		err := loader.LoadUserItems(ctx, exec, queryMods...)

		// Don't cause an issue due to missing relationships
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}

		return err
	})
}

// LoadUserItems loads the user's Items into the .R struct
func (o *User) LoadUserItems(ctx context.Context, exec bob.Executor, mods ...bob.Mod[*dialect.SelectQuery]) error {
	if o == nil {
		return nil
	}

	// Reset the relationship
	o.R.Items = nil

	related, err := o.Items(mods...).All(ctx, exec)
	if err != nil {
		return err
	}

	for _, rel := range related {
		rel.R.User = o
	}

	o.R.Items = related
	return nil
}

// LoadUserItems loads the user's Items into the .R struct
func (os UserSlice) LoadUserItems(ctx context.Context, exec bob.Executor, mods ...bob.Mod[*dialect.SelectQuery]) error {
	if len(os) == 0 {
		return nil
	}

	items, err := os.Items(mods...).All(ctx, exec)
	if err != nil {
		return err
	}

	for _, o := range os {
		o.R.Items = nil
	}

	for _, o := range os {
		for _, rel := range items {
			if o.ID != rel.UserID {
				continue
			}

			rel.R.User = o

			o.R.Items = append(o.R.Items, rel)
		}
	}

	return nil
}

func PreloadUserProfilePictureFile(opts ...sqlite.PreloadOption) sqlite.Preloader {
	return sqlite.Preload[*File, FileSlice](orm.Relationship{
		Name: "ProfilePictureFile",
		Sides: []orm.RelSide{
			{
				From: TableNames.Users,
				To:   TableNames.Files,
				FromColumns: []string{
					ColumnNames.Users.ProfilePictureID,
				},
				ToColumns: []string{
					ColumnNames.Files.ID,
				},
			},
		},
	}, Files.Columns().Names(), opts...)
}

func ThenLoadUserProfilePictureFile(queryMods ...bob.Mod[*dialect.SelectQuery]) sqlite.Loader {
	return sqlite.Loader(func(ctx context.Context, exec bob.Executor, retrieved any) error {
		loader, isLoader := retrieved.(interface {
			LoadUserProfilePictureFile(context.Context, bob.Executor, ...bob.Mod[*dialect.SelectQuery]) error
		})
		if !isLoader {
			return fmt.Errorf("object %T cannot load UserProfilePictureFile", retrieved)
		}

		err := loader.LoadUserProfilePictureFile(ctx, exec, queryMods...)

		// Don't cause an issue due to missing relationships
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}

		return err
	})
}

// LoadUserProfilePictureFile loads the user's ProfilePictureFile into the .R struct
func (o *User) LoadUserProfilePictureFile(ctx context.Context, exec bob.Executor, mods ...bob.Mod[*dialect.SelectQuery]) error {
	if o == nil {
		return nil
	}

	// Reset the relationship
	o.R.ProfilePictureFile = nil

	related, err := o.ProfilePictureFile(mods...).One(ctx, exec)
	if err != nil {
		return err
	}

	related.R.ProfilePictureUsers = UserSlice{o}

	o.R.ProfilePictureFile = related
	return nil
}

// LoadUserProfilePictureFile loads the user's ProfilePictureFile into the .R struct
func (os UserSlice) LoadUserProfilePictureFile(ctx context.Context, exec bob.Executor, mods ...bob.Mod[*dialect.SelectQuery]) error {
	if len(os) == 0 {
		return nil
	}

	files, err := os.ProfilePictureFile(mods...).All(ctx, exec)
	if err != nil {
		return err
	}

	for _, o := range os {
		for _, rel := range files {
			if o.ProfilePictureID.GetOrZero() != rel.ID {
				continue
			}

			rel.R.ProfilePictureUsers = append(rel.R.ProfilePictureUsers, o)

			o.R.ProfilePictureFile = rel
			break
		}
	}

	return nil
}

func insertUserFiles0(ctx context.Context, exec bob.Executor, files1 []*FileSetter, user0 *User) (FileSlice, error) {
	for i := range files1 {
		files1[i].UserID = omit.From(user0.ID)
	}

	ret, err := Files.Insert(bob.ToMods(files1...)).All(ctx, exec)
	if err != nil {
		return ret, fmt.Errorf("insertUserFiles0: %w", err)
	}

	return ret, nil
}

func attachUserFiles0(ctx context.Context, exec bob.Executor, count int, files1 FileSlice, user0 *User) (FileSlice, error) {
	setter := &FileSetter{
		UserID: omit.From(user0.ID),
	}

	err := files1.UpdateAll(ctx, exec, *setter)
	if err != nil {
		return nil, fmt.Errorf("attachUserFiles0: %w", err)
	}

	return files1, nil
}

func (user0 *User) InsertFiles(ctx context.Context, exec bob.Executor, related ...*FileSetter) error {
	if len(related) == 0 {
		return nil
	}

	var err error

	files1, err := insertUserFiles0(ctx, exec, related, user0)
	if err != nil {
		return err
	}

	user0.R.Files = append(user0.R.Files, files1...)

	for _, rel := range files1 {
		rel.R.User = user0
	}
	return nil
}

func (user0 *User) AttachFiles(ctx context.Context, exec bob.Executor, related ...*File) error {
	if len(related) == 0 {
		return nil
	}

	var err error
	files1 := FileSlice(related)

	_, err = attachUserFiles0(ctx, exec, len(related), files1, user0)
	if err != nil {
		return err
	}

	user0.R.Files = append(user0.R.Files, files1...)

	for _, rel := range related {
		rel.R.User = user0
	}

	return nil
}

func insertUserItems0(ctx context.Context, exec bob.Executor, items1 []*ItemSetter, user0 *User) (ItemSlice, error) {
	for i := range items1 {
		items1[i].UserID = omit.From(user0.ID)
	}

	ret, err := Items.Insert(bob.ToMods(items1...)).All(ctx, exec)
	if err != nil {
		return ret, fmt.Errorf("insertUserItems0: %w", err)
	}

	return ret, nil
}

func attachUserItems0(ctx context.Context, exec bob.Executor, count int, items1 ItemSlice, user0 *User) (ItemSlice, error) {
	setter := &ItemSetter{
		UserID: omit.From(user0.ID),
	}

	err := items1.UpdateAll(ctx, exec, *setter)
	if err != nil {
		return nil, fmt.Errorf("attachUserItems0: %w", err)
	}

	return items1, nil
}

func (user0 *User) InsertItems(ctx context.Context, exec bob.Executor, related ...*ItemSetter) error {
	if len(related) == 0 {
		return nil
	}

	var err error

	items1, err := insertUserItems0(ctx, exec, related, user0)
	if err != nil {
		return err
	}

	user0.R.Items = append(user0.R.Items, items1...)

	for _, rel := range items1 {
		rel.R.User = user0
	}
	return nil
}

func (user0 *User) AttachItems(ctx context.Context, exec bob.Executor, related ...*Item) error {
	if len(related) == 0 {
		return nil
	}

	var err error
	items1 := ItemSlice(related)

	_, err = attachUserItems0(ctx, exec, len(related), items1, user0)
	if err != nil {
		return err
	}

	user0.R.Items = append(user0.R.Items, items1...)

	for _, rel := range related {
		rel.R.User = user0
	}

	return nil
}

func attachUserProfilePictureFile0(ctx context.Context, exec bob.Executor, count int, user0 *User, file1 *File) (*User, error) {
	setter := &UserSetter{
		ProfilePictureID: omitnull.From(file1.ID),
	}

	err := user0.Update(ctx, exec, setter)
	if err != nil {
		return nil, fmt.Errorf("attachUserProfilePictureFile0: %w", err)
	}

	return user0, nil
}

func (user0 *User) InsertProfilePictureFile(ctx context.Context, exec bob.Executor, related *FileSetter) error {
	file1, err := Files.Insert(related).One(ctx, exec)
	if err != nil {
		return fmt.Errorf("inserting related objects: %w", err)
	}

	_, err = attachUserProfilePictureFile0(ctx, exec, 1, user0, file1)
	if err != nil {
		return err
	}

	user0.R.ProfilePictureFile = file1

	file1.R.ProfilePictureUsers = append(file1.R.ProfilePictureUsers, user0)

	return nil
}

func (user0 *User) AttachProfilePictureFile(ctx context.Context, exec bob.Executor, file1 *File) error {
	var err error

	_, err = attachUserProfilePictureFile0(ctx, exec, 1, user0, file1)
	if err != nil {
		return err
	}

	user0.R.ProfilePictureFile = file1

	file1.R.ProfilePictureUsers = append(file1.R.ProfilePictureUsers, user0)

	return nil
}
