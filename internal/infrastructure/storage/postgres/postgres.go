package postgres

import (
	"context"

	"cqrs-layout-example/config"
	"cqrs-layout-example/internal/domain/entity"
	"cqrs-layout-example/internal/infrastructure/storage/dto"

	"github.com/NikitaTsaralov/utils/utils/trace"
	txManager "github.com/avito-tech/go-transaction-manager/sqlx"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel"
)

type Storage struct {
	cfg       *config.Config
	db        *sqlx.DB
	ctxGetter *txManager.CtxGetter
}

func NewStorage(cfg *config.Config, db *sqlx.DB, ctxGetter *txManager.CtxGetter) *Storage {
	return &Storage{
		cfg:       cfg,
		db:        db,
		ctxGetter: ctxGetter,
	}
}

func (s *Storage) Create(ctx context.Context, entity *entity.Entity) (int64, error) {
	ctx, span := otel.Tracer("").Start(ctx, "pgrepo.Entity.Create")
	defer span.End()

	var id int64

	dtoEntity := dto.NewEntityFromDomain(entity)

	err := s.ctxGetter.DefaultTrOrDB(ctx, s.db).GetContext(ctx, &id, queryCreate, dtoEntity)
	if err != nil {
		return 0, trace.Wrapf(span, err, "pgrepo.Entity.Create.GetContext(args: %#v)", entity)
	}

	return id, nil
}

func (s *Storage) Update(ctx context.Context, entity *entity.Entity) error {
	ctx, span := otel.Tracer("").Start(ctx, "pgrepo.Entity.Update")
	defer span.End()

	err := s.ctxGetter.DefaultTrOrDB(ctx, s.db).GetContext(ctx, nil, queryUpdate)
	if err != nil {
		return trace.Wrapf(span, err, "pgrepo.Entity.Update.GetContext(args: %#v)", entity)
	}

	return nil
}

func (s *Storage) GetByID(ctx context.Context, id uuid.UUID) (*entity.Entity, error) {
	ctx, span := otel.Tracer("").Start(ctx, "pgrepo.Entity.GetByID")
	defer span.End()

	var entity dto.Entity

	err := s.db.GetContext(ctx, &entity, queryGetByID, id)
	if err != nil {
		return nil, trace.Wrapf(span, err, "pgrepo.Entity.GetByID.GetContext(id: %s)", id.String())
	}

	return entity.ToDomain(), nil
}

func (s *Storage) Fetch(ctx context.Context) (entity.Entities, error) {
	ctx, span := otel.Tracer("").Start(ctx, "pgrepo.Entity.Fetch")
	defer span.End()

	var entities dto.Entities

	err := s.db.SelectContext(ctx, &entities, queryFetch)
	if err != nil {
		return nil, trace.Wrapf(span, err, "pgrepo.Entity.Fetch.SelectContext()")
	}

	return entities.ToDomain(), nil
}
