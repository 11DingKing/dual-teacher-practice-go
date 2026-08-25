package domain

import "time"

type Role string

const (
	RoleTeacher    Role = "teacher"
	Role学院         Role = "college"
	RoleEnterprise Role = "enterprise"
	RoleHR         Role = "hr"
	RoleAdmin      Role = "admin"
)

type ApplicationStatus string

const (
	StatusDraft           ApplicationStatus = "draft"
	StatusSubmitted       ApplicationStatus = "submitted"
	StatusCollegeApproved ApplicationStatus = "college_approved"
	StatusEvidenceReview  ApplicationStatus = "evidence_review"
	StatusPeerReview      ApplicationStatus = "peer_review"
	StatusCertified       ApplicationStatus = "certified"
	StatusRenewal         ApplicationStatus = "renewal"
	StatusRejected        ApplicationStatus = "rejected"
	StatusExpired         ApplicationStatus = "expired"
)

type EvidenceStatus string

const (
	EvidencePending    EvidenceStatus = "pending"
	EvidenceVerified   EvidenceStatus = "verified"
	EvidenceCorrection EvidenceStatus = "correction"
	EvidenceWithdrawn  EvidenceStatus = "withdrawn"
	EvidenceDisputed   EvidenceStatus = "disputed"
)

type ReviewDecision string

const (
	ReviewPass    ReviewDecision = "pass"
	ReviewFail    ReviewDecision = "fail"
	ReviewPending ReviewDecision = "pending"
)

type User struct {
	ID, Username, DisplayName string
	Role                      Role
	RevokedAt                 *time.Time
	CreatedAt                 time.Time
}
type Session struct {
	ID, UserID string
	ExpiresAt  time.Time
	RevokedAt  *time.Time
	CreatedAt  time.Time
}
type Quota struct {
	ID, Name                                                      string
	MaxHours, UsedHours, MaxBudgetCents, UsedBudgetCents, Version int
	Active                                                        bool
}
type Application struct {
	ID, TeacherID, QuotaID, IdempotencyKey        string
	Year                                          int
	Status                                        ApplicationStatus
	RequestedHours, RequestedBudgetCents, Version int
	DueAt, CreatedAt, UpdatedAt                   time.Time
}
type Practice struct {
	ID, ApplicationID, CompanyName, MentorName, ContactEmail string
	Status                                                   string
	PlannedHours, ActualHours, Version                       int
	StartsAt, EndsAt                                         time.Time
}
type Evidence struct {
	ID, ApplicationID, Kind, Title, URI string
	Status                              EvidenceStatus
	SubmittedBy, VerifiedBy             string
	VerifiedAt, WithdrawnAt             *time.Time
	DisputeNote                         string
	CreatedAt                           time.Time
}
type Review struct {
	ID, ApplicationID, ReviewerID string
	ReviewerRole                  Role
	Score                         int
	Decision                      ReviewDecision
	Comment                       string
	CreatedAt                     time.Time
}
type AuditEvent struct {
	ID, ActorID, Action, EntityType, EntityID, Outcome, RequestID, Details string
	CreatedAt                                                              time.Time
}
type WorkerJob struct {
	ID, JobType, Payload, Status, LastError string
	Attempts                                int
	NextRunAt, CreatedAt, UpdatedAt         time.Time
}
