package business

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	blogdomain "sitehub/service/blog/domain"
	blogrepository "sitehub/service/blog/model"
	blogentity "sitehub/service/blog/model/entity"
)

// Each scheduled publication locks the same parent as manual actions and publishes the frozen revision.
func (s *articleService) publishDue(before time.Time) (int64, error) {
	scheduled, err := s.articleRepo.GetScheduledArticles(time.Time{})
	if err != nil {
		return 0, err
	}
	var count int64
	var failures []error
	for _, candidate := range scheduled {
		if candidate.ScheduledAt == nil || candidate.ScheduledAt.After(before) {
			continue
		}
		changed := false
		err = s.articleRepo.WorkflowTransaction(func(repo blogrepository.ArticleRepository) error {
			row, e := repo.WorkflowGet(candidate.ID, true)
			if e != nil {
				return e
			}
			if row.DeletedAt != nil || row.Status != 2 || row.ScheduledAt == nil || row.ScheduledAt.After(before) {
				return nil
			}
			var publication blogdomain.ArticleView
			if row.ScheduledVersion > 0 {
				rev, e := repo.RevisionByVersion(row.ID, row.ScheduledVersion)
				if e != nil {
					return e
				}
				if e = json.Unmarshal([]byte(rev.Payload), &publication); e != nil {
					return e
				}
			} else {
				publication = snapshot(row)
			}
			if publication.Title == "" || !hasArticleContent(publication.Content) {
				return blogdomain.ErrInvalidArticle
			}
			if e = repo.ValidateReferences(publication.CategoryID, publication.TagIDs, assetIDs(publication.ArticleInput), row.UpdatedBy, row.ID); e != nil {
				return e
			}
			when := *row.ScheduledAt
			publication.Status = 1
			publication.PublishedAt = &when
			publication.ScheduledAt = nil
			publication.HasUnpublishedChanges = false
			raw, e := json.Marshal(publication)
			if e != nil {
				return e
			}
			if e = repo.Publish(&blogentity.ArticlePublication{ArticleID: row.ID, Title: publication.Title, Slug: publication.Slug, Content: publication.Content, CategoryID: publication.CategoryID, Payload: string(raw), PublishedAt: when}, publication.TagIDs); e != nil {
				return e
			}
			row.Status = 1
			row.PublishedVersion = publication.Version
			row.PublishedAt = &when
			row.ScheduledAt = nil
			row.ScheduledVersion = 0
			row.Version++
			row.UpdatedAt = before
			if e = repo.WorkflowSave(row); e != nil {
				return e
			}
			if _, e = recordRevision(repo, row, 0, "scheduled-publish"); e != nil {
				return e
			}
			changed = true
			return nil
		})
		if err != nil && !errors.Is(err, blogdomain.ErrArticleNotFound) {
			// One invalid article must not prevent unrelated due articles from publishing.
			failures = append(failures, fmt.Errorf("scheduled article %d: %w", candidate.ID, err))
		}
		if changed {
			count++
		}
	}
	return count, errors.Join(failures...)
}
