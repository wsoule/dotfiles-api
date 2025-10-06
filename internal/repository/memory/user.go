package memory

import (
	"context"
	"sort"
	"sync"

	"dotfiles-web/internal/models"
	"dotfiles-web/pkg/errors"
)

type UserRepository struct {
	users     map[string]*models.User
	favorites map[string][]string
	mutex     sync.RWMutex
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		users:     make(map[string]*models.User),
		favorites: make(map[string][]string),
	}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.users[user.ID]; exists {
		return errors.NewConflictError("user already exists")
	}

	for _, existingUser := range r.users {
		if existingUser.Username == user.Username {
			return errors.NewConflictError("username already taken")
		}
		if existingUser.Email == user.Email {
			return errors.NewConflictError("email already taken")
		}
	}

	r.users[user.ID] = user
	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return nil, errors.NewNotFoundError("user")
	}

	return user, nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, user := range r.users {
		if user.Username == username {
			return user, nil
		}
	}

	return nil, errors.NewNotFoundError("user")
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, user := range r.users {
		if user.Email == email {
			return user, nil
		}
	}

	return nil, errors.NewNotFoundError("user")
}

func (r *UserRepository) GetByGitHubID(ctx context.Context, githubID int) (*models.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, user := range r.users {
		if user.GitHubID == githubID {
			return user, nil
		}
	}

	return nil, errors.NewNotFoundError("user")
}

func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.users[user.ID]; !exists {
		return errors.NewNotFoundError("user")
	}

	for id, existingUser := range r.users {
		if id != user.ID {
			if existingUser.Username == user.Username {
				return errors.NewConflictError("username already taken")
			}
			if existingUser.Email == user.Email {
				return errors.NewConflictError("email already taken")
			}
		}
	}

	r.users[user.ID] = user
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.users[id]; !exists {
		return errors.NewNotFoundError("user")
	}

	delete(r.users, id)
	delete(r.favorites, id)
	return nil
}

func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]*models.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	users := make([]*models.User, 0, len(r.users))
	for _, user := range r.users {
		users = append(users, user)
	}

	sort.Slice(users, func(i, j int) bool {
		return users[i].CreatedAt.After(users[j].CreatedAt)
	})

	start := offset
	if start > len(users) {
		return []*models.User{}, nil
	}

	end := start + limit
	if end > len(users) {
		end = len(users)
	}

	return users[start:end], nil
}

func (r *UserRepository) AddFavorite(ctx context.Context, userID, templateID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.users[userID]; !exists {
		return errors.NewNotFoundError("user")
	}

	favorites := r.favorites[userID]
	for _, fav := range favorites {
		if fav == templateID {
			return errors.NewConflictError("template already in favorites")
		}
	}

	r.favorites[userID] = append(favorites, templateID)
	return nil
}

func (r *UserRepository) RemoveFavorite(ctx context.Context, userID, templateID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.users[userID]; !exists {
		return errors.NewNotFoundError("user")
	}

	favorites := r.favorites[userID]
	for i, fav := range favorites {
		if fav == templateID {
			r.favorites[userID] = append(favorites[:i], favorites[i+1:]...)
			return nil
		}
	}

	return errors.NewNotFoundError("favorite")
}

func (r *UserRepository) GetFavorites(ctx context.Context, userID string) ([]string, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if _, exists := r.users[userID]; !exists {
		return nil, errors.NewNotFoundError("user")
	}

	favorites := r.favorites[userID]
	if favorites == nil {
		return []string{}, nil
	}

	result := make([]string, len(favorites))
	copy(result, favorites)
	return result, nil
}