package mysql

import (
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	productdomain "sitehub/service/product/domain"
)

func TestDocumentLookupsUseSeparateVersionScopedKeys(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	orm, err := gorm.Open(mysql.New(mysql.Config{Conn: db, SkipInitializeWithVersion: true}), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	repo := NewProductDocsRepository(orm)
	for _, tc := range []struct {
		field string
		read  func(string, string) (productdomain.ProductDocument, error)
	}{{"doc_key", repo.Document}, {"slug", repo.DocumentBySlug}} {
		query := regexp.QuoteMeta("SELECT * FROM `product_documents` WHERE version_id = ? AND " + tc.field + " = ? LIMIT ?")
		mock.ExpectQuery(query).WithArgs("v1", "quick-start", 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "doc_key", "slug"}).AddRow(1, "quick-start", "install"))
		if _, err = tc.read("v1", "quick-start"); err != nil {
			t.Fatal(err)
		}
		mock.ExpectQuery(query).WithArgs("v2", "missing", 1).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
		if _, err = tc.read("v2", "missing"); !errors.Is(err, productdomain.ErrDocumentNotFound) {
			t.Fatalf("missing sentinel lost: %v", err)
		}
		cause := errors.New("connection lost")
		mock.ExpectQuery(query).WithArgs("v1", "quick-start", 1).WillReturnError(cause)
		if _, err = tc.read("v1", "quick-start"); !errors.Is(err, cause) {
			t.Fatalf("driver cause lost: %v", err)
		}
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
