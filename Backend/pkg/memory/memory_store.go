package memory

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/elvisouma/salmar-ai/pkg/models"
)

// InMemoryStore provides a simple in-memory implementation of the memory store
// For production, this would be replaced with a vector database like Pinecone or Redis
type InMemoryStore struct {
	memories map[string][]models.MemoryEntry // key: userID:conversationID
	mu       sync.RWMutex
}

// NewInMemoryStore creates a new in-memory store
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		memories: make(map[string][]models.MemoryEntry),
	}
}

// StoreMemory stores a memory entry
func (s *InMemoryStore) StoreMemory(ctx context.Context, memory *models.MemoryEntry) error {
	if memory == nil {
		return errors.New("memory cannot be nil")
	}

	if memory.UserID == "" || memory.ConversationID == "" {
		return errors.New("userID and conversationID are required")
	}

	// Set timestamp if not provided
	if memory.Timestamp == 0 {
		memory.Timestamp = time.Now().Unix()
	}

	key := memory.UserID + ":" + memory.ConversationID

	s.mu.Lock()
	defer s.mu.Unlock()

	// Add memory to the store
	s.memories[key] = append(s.memories[key], *memory)

	return nil
}

// RetrieveMemories retrieves memories for a user and conversation
func (s *InMemoryStore) RetrieveMemories(ctx context.Context, userID, conversationID string, limit int) ([]models.MemoryEntry, error) {
	if userID == "" || conversationID == "" {
		return nil, errors.New("userID and conversationID are required")
	}

	key := userID + ":" + conversationID

	s.mu.RLock()
	defer s.mu.RUnlock()

	memories, exists := s.memories[key]
	if !exists {
		return []models.MemoryEntry{}, nil
	}

	// Return the most recent memories up to the limit
	if len(memories) <= limit {
		return memories, nil
	}

	return memories[len(memories)-limit:], nil
}

// SearchSimilarMemories finds memories similar to the query
// This is a simplified implementation - in production, this would use vector similarity search
func (s *InMemoryStore) SearchSimilarMemories(ctx context.Context, query string, limit int) ([]models.MemoryEntry, error) {
	if query == "" {
		return nil, errors.New("query cannot be empty")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Collect all memories
	var allMemories []models.MemoryEntry
	for _, memories := range s.memories {
		allMemories = append(allMemories, memories...)
	}

	// In a real implementation, this would perform vector similarity search
	// For now, just return the most recent memories
	if len(allMemories) <= limit {
		return allMemories, nil
	}

	return allMemories[len(allMemories)-limit:], nil
}

// VectorMemoryStore is a placeholder for a production-ready vector database implementation
// This would be implemented with a real vector database like Pinecone, Redis, or similar
type VectorMemoryStore struct {
	// Connection to vector database would be here
}

// NewVectorMemoryStore creates a new vector memory store
func NewVectorMemoryStore(connectionString string) (*VectorMemoryStore, error) {
	// Initialize connection to vector database
	return &VectorMemoryStore{}, nil
}

// StoreMemory stores a memory entry with vector embedding
func (s *VectorMemoryStore) StoreMemory(ctx context.Context, memory *models.MemoryEntry) error {
	// Implementation would store the memory with its embedding in the vector database
	return errors.New("not implemented")
}

// RetrieveMemories retrieves memories for a user and conversation
func (s *VectorMemoryStore) RetrieveMemories(ctx context.Context, userID, conversationID string, limit int) ([]models.MemoryEntry, error) {
	// Implementation would retrieve memories from the vector database
	return nil, errors.New("not implemented")
}

// SearchSimilarMemories finds memories similar to the query using vector similarity
func (s *VectorMemoryStore) SearchSimilarMemories(ctx context.Context, query string, limit int) ([]models.MemoryEntry, error) {
	// Implementation would perform vector similarity search
	return nil, errors.New("not implemented")
}
