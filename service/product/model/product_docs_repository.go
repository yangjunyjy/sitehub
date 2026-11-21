package repository

import (
	productdomain "sitehub/service/product/domain"
)

// Transaction 的回调只能使用传入仓储，确保锁、校验与写入使用同一连接。
type ProductDocsRepository interface {
	Transaction(func(ProductDocsRepository) error) error
	Site(lock bool) (productdomain.DocSite, error)
	Modes() ([]productdomain.DocMode, error)
	PutMode(productdomain.DocMode) error
	DeleteMode(string) error
	Versions() ([]productdomain.DocVersion, error)
	Groups(version string) ([]productdomain.DocGroup, error)
	Documents(version string, body bool) ([]productdomain.ProductDocument, error)
	Document(version, key string) (productdomain.ProductDocument, error)
	DocumentBySlug(version, slug string) (productdomain.ProductDocument, error)
	PublishedDocument(version, key string) (productdomain.ProductDocument, error)
	PublishedDocuments(version string) ([]productdomain.ProductDocument, error)
	DocumentID(version, key string) (uint64, error)
	History(documentID uint64) ([]productdomain.DocumentRevision, error)
	Historical(documentID, historyID uint64) (productdomain.DocumentRevision, error)
	PutSite(productdomain.DocSite) error
	PutVersion(productdomain.DocVersion) error
	PutGroup(*productdomain.DocGroup) error
	PutDocument(*productdomain.ProductDocument) error
	AddRevision(*productdomain.DocumentRevision) error
	DeleteVersion(string) error
	DeleteGroup(uint64) error
	DeleteDocument(uint64) error
}
