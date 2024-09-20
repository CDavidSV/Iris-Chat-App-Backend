package models

import (
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Channel struct {
	ChannelID   string
	ChannelName string
	OwnerID     string
	ChannelType string
	Description NullString
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ChannelMember struct {
	ChannelID string
	UserID    string
	JoinedAt  time.Time
	IsAdmnin  bool
	Hidden    bool
}

type ChannelModel struct {
	DB *pgxpool.Pool
}

func (m *ChannelModel) CreateChannel() {

}

func (m *ChannelModel) DeleteChannel() {

}

func (m *ChannelModel) AddMember() {

}

func (m *ChannelModel) RemoveMember() {

}

func (m *ChannelModel) FetchChannel() {

}

func (m *ChannelModel) FetchChannels() {

}

func (m *ChannelModel) FetchMembers() {

}

func (m *ChannelModel) ChangeAdminPerms() {

}

func (m *ChannelModel) ChangeChannelHiddenState() {

}
