import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
export type Maybe<T> = T | null;
export type InputMaybe<T> = Maybe<T>;
export type Exact<T extends { [key: string]: unknown }> = { [K in keyof T]: T[K] };
export type MakeOptional<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]?: Maybe<T[SubKey]> };
export type MakeMaybe<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]: Maybe<T[SubKey]> };
export type MakeEmpty<T extends { [key: string]: unknown }, K extends keyof T> = { [_ in K]?: never };
export type Incremental<T> = T | { [P in keyof T]?: P extends ' $fragmentName' | '__typename' ? T[P] : never };
const defaultOptions = {} as const;
/** All built-in and custom scalars, mapped to their actual values */
export type Scalars = {
  ID: { input: string; output: string; }
  String: { input: string; output: string; }
  Boolean: { input: boolean; output: boolean; }
  Int: { input: number; output: number; }
  Float: { input: number; output: number; }
  /** An RFC 3339 timestamp. */
  DateTime: { input: string; output: string; }
};

export type AddAssumptionInput = {
  confidence: Scalars['Float']['input'];
  decisionId: Scalars['ID']['input'];
  statement: Scalars['String']['input'];
};

export type AddTeamMemberInput = {
  role: Role;
  teamId: Scalars['ID']['input'];
  userId: Scalars['ID']['input'];
};

export type AddTeamMemberPayload = {
  __typename?: 'AddTeamMemberPayload';
  membership?: Maybe<Membership>;
  userErrors: Array<UserError>;
};

/** An assumption that underpins a decision, with a confidence in [0, 1]. */
export type Assumption = {
  __typename?: 'Assumption';
  confidence: Scalars['Float']['output'];
  createdAt: Scalars['DateTime']['output'];
  decision: Decision;
  evidence: Array<Evidence>;
  id: Scalars['ID']['output'];
  statement: Scalars['String']['output'];
  updatedAt: Scalars['DateTime']['output'];
};

export type AssumptionPayload = {
  __typename?: 'AssumptionPayload';
  assumption?: Maybe<Assumption>;
  userErrors: Array<UserError>;
};

export type AttachEvidenceInput = {
  assumptionId: Scalars['ID']['input'];
  content: Scalars['String']['input'];
  sourceType: EvidenceSourceType;
  sourceUrl?: InputMaybe<Scalars['String']['input']>;
};

/** The result of an authentication mutation: tokens plus the authenticated user. */
export type AuthPayload = {
  __typename?: 'AuthPayload';
  accessExpiresAt?: Maybe<Scalars['DateTime']['output']>;
  accessToken?: Maybe<Scalars['String']['output']>;
  refreshToken?: Maybe<Scalars['String']['output']>;
  user?: Maybe<User>;
  userErrors: Array<UserError>;
};

/** The result of a bias-detection request. */
export type BiasReport = {
  __typename?: 'BiasReport';
  biases: Array<DetectedBias>;
  summary: Scalars['String']['output'];
};

/** One decile of a calibration (reliability) curve. */
export type CalibrationBin = {
  __typename?: 'CalibrationBin';
  /** The decile index, 1..10, of predicted probability. */
  bucket: Scalars['Int']['output'];
  /** Average predicted probability of forecasts in the bin. */
  meanPredicted: Scalars['Float']['output'];
  /** Fraction of those forecasts that actually succeeded, in [0, 1]. */
  observedFrequency: Scalars['Float']['output'];
  /** Number of forecasts in the bin. */
  sampleSize: Scalars['Int']['output'];
};

/** The full set of calibration bins for a team, ordered by bucket. */
export type CalibrationReport = {
  __typename?: 'CalibrationReport';
  bins: Array<CalibrationBin>;
};

export type ChangeMemberRoleInput = {
  role: Role;
  teamId: Scalars['ID']['input'];
  userId: Scalars['ID']['input'];
};

export type ChangeMemberRolePayload = {
  __typename?: 'ChangeMemberRolePayload';
  membership?: Maybe<Membership>;
  userErrors: Array<UserError>;
};

export type CreateDecisionInput = {
  description: Scalars['String']['input'];
  teamId: Scalars['ID']['input'];
  title: Scalars['String']['input'];
};

export type CreatePredictionInput = {
  decisionId: Scalars['ID']['input'];
  probability: Scalars['Float']['input'];
  resolvesAt?: InputMaybe<Scalars['DateTime']['input']>;
  statement: Scalars['String']['input'];
};

export type CreateTeamInput = {
  name: Scalars['String']['input'];
};

export type CreateTeamPayload = {
  __typename?: 'CreateTeamPayload';
  team?: Maybe<Team>;
  userErrors: Array<UserError>;
};

/** A decision: the root aggregate that assumptions, predictions, and an outcome hang from. */
export type Decision = {
  __typename?: 'Decision';
  assumptions: Array<Assumption>;
  createdAt: Scalars['DateTime']['output'];
  decidedAt?: Maybe<Scalars['DateTime']['output']>;
  description: Scalars['String']['output'];
  id: Scalars['ID']['output'];
  outcome?: Maybe<Outcome>;
  owner: User;
  predictions: Array<Prediction>;
  status: DecisionStatus;
  team: Team;
  title: Scalars['String']['output'];
  updatedAt: Scalars['DateTime']['output'];
};

export type DecisionConnection = {
  __typename?: 'DecisionConnection';
  edges: Array<DecisionEdge>;
  pageInfo: PageInfo;
  totalCount: Scalars['Int']['output'];
};

export type DecisionEdge = {
  __typename?: 'DecisionEdge';
  cursor: Scalars['String']['output'];
  node: Decision;
};

export type DecisionPayload = {
  __typename?: 'DecisionPayload';
  decision?: Maybe<Decision>;
  userErrors: Array<UserError>;
};

/** Lifecycle status of a decision. */
export type DecisionStatus =
  | 'ACTIVE'
  | 'ARCHIVED'
  | 'DECIDED'
  | 'DRAFT';

/** Shared payload for mutations that delete an entity. */
export type DeletePayload = {
  __typename?: 'DeletePayload';
  success: Scalars['Boolean']['output'];
  userErrors: Array<UserError>;
};

/** A single cognitive bias flagged in a piece of text. */
export type DetectedBias = {
  __typename?: 'DetectedBias';
  explanation: Scalars['String']['output'];
  name: Scalars['String']['output'];
};

/** Evidence attached to an assumption that supports or challenges it. */
export type Evidence = {
  __typename?: 'Evidence';
  assumption: Assumption;
  content: Scalars['String']['output'];
  createdAt: Scalars['DateTime']['output'];
  id: Scalars['ID']['output'];
  sourceType: EvidenceSourceType;
  sourceUrl?: Maybe<Scalars['String']['output']>;
  updatedAt: Scalars['DateTime']['output'];
};

export type EvidencePayload = {
  __typename?: 'EvidencePayload';
  evidence?: Maybe<Evidence>;
  userErrors: Array<UserError>;
};

/** The kind of source a piece of evidence comes from. */
export type EvidenceSourceType =
  | 'DATASET'
  | 'DOCUMENT'
  | 'NOTE'
  | 'URL';

export type LoginInput = {
  email: Scalars['String']['input'];
  password: Scalars['String']['input'];
};

/** A user's membership of a team and the role they hold within it. */
export type Membership = {
  __typename?: 'Membership';
  createdAt: Scalars['DateTime']['output'];
  role: Role;
  team: Team;
  user: User;
};

export type Mutation = {
  __typename?: 'Mutation';
  /** Add an assumption to a decision. */
  addAssumption: AssumptionPayload;
  /** Add a user to a team with a role. */
  addTeamMember: AddTeamMemberPayload;
  /** Attach evidence to an assumption. */
  attachEvidence: EvidencePayload;
  /** Change a member's role within a team. */
  changeMemberRole: ChangeMemberRolePayload;
  /** Create a decision within a team. */
  createDecision: DecisionPayload;
  /** Create a forecast for a decision. */
  createPrediction: PredictionPayload;
  /** Create a team owned by the authenticated user. */
  createTeam: CreateTeamPayload;
  /** Critique an assumption; null when AI assistance is disabled. */
  critiqueAssumption?: Maybe<Scalars['String']['output']>;
  /** Detect cognitive biases in text; null when AI assistance is disabled. */
  detectBias?: Maybe<BiasReport>;
  /** Authenticate with email and password. */
  login: AuthPayload;
  /** Invalidate a refresh token. Web clients may omit the argument and rely on the httpOnly refresh cookie. */
  logout: Scalars['Boolean']['output'];
  /** Record (or update) the outcome of a decision. */
  recordOutcome: OutcomePayload;
  /** Exchange a refresh token for a new access token. Web clients may omit the argument and rely on the httpOnly refresh cookie with CSRF protection. */
  refreshToken: AuthPayload;
  /** Register a new account. */
  register: AuthPayload;
  /** Remove an assumption. */
  removeAssumption: DeletePayload;
  /** Remove a piece of evidence. */
  removeEvidence: DeletePayload;
  /** Remove a member from a team. */
  removeTeamMember: RemoveTeamMemberPayload;
  /** Summarize a piece of evidence; null when AI assistance is disabled. */
  summarizeEvidence?: Maybe<Scalars['String']['output']>;
  /** Move a decision to a new lifecycle status. */
  transitionDecision: DecisionPayload;
  /** Update an assumption's statement and confidence. */
  updateAssumption: AssumptionPayload;
  /** Update a decision's title and description. */
  updateDecision: DecisionPayload;
  /** Update a piece of evidence. */
  updateEvidence: EvidencePayload;
  /** Update a forecast. */
  updatePrediction: PredictionPayload;
  /** Update the authenticated user's profile. */
  updateProfile: UpdateProfilePayload;
};


export type MutationAddAssumptionArgs = {
  input: AddAssumptionInput;
};


export type MutationAddTeamMemberArgs = {
  input: AddTeamMemberInput;
};


export type MutationAttachEvidenceArgs = {
  input: AttachEvidenceInput;
};


export type MutationChangeMemberRoleArgs = {
  input: ChangeMemberRoleInput;
};


export type MutationCreateDecisionArgs = {
  input: CreateDecisionInput;
};


export type MutationCreatePredictionArgs = {
  input: CreatePredictionInput;
};


export type MutationCreateTeamArgs = {
  input: CreateTeamInput;
};


export type MutationCritiqueAssumptionArgs = {
  assumptionId: Scalars['ID']['input'];
};


export type MutationDetectBiasArgs = {
  text: Scalars['String']['input'];
};


export type MutationLoginArgs = {
  input: LoginInput;
};


export type MutationLogoutArgs = {
  token?: InputMaybe<Scalars['String']['input']>;
};


export type MutationRecordOutcomeArgs = {
  input: RecordOutcomeInput;
};


export type MutationRefreshTokenArgs = {
  token?: InputMaybe<Scalars['String']['input']>;
};


export type MutationRegisterArgs = {
  input: RegisterInput;
};


export type MutationRemoveAssumptionArgs = {
  id: Scalars['ID']['input'];
};


export type MutationRemoveEvidenceArgs = {
  id: Scalars['ID']['input'];
};


export type MutationRemoveTeamMemberArgs = {
  input: RemoveTeamMemberInput;
};


export type MutationSummarizeEvidenceArgs = {
  evidenceId: Scalars['ID']['input'];
};


export type MutationTransitionDecisionArgs = {
  input: TransitionDecisionInput;
};


export type MutationUpdateAssumptionArgs = {
  input: UpdateAssumptionInput;
};


export type MutationUpdateDecisionArgs = {
  input: UpdateDecisionInput;
};


export type MutationUpdateEvidenceArgs = {
  input: UpdateEvidenceInput;
};


export type MutationUpdatePredictionArgs = {
  input: UpdatePredictionInput;
};


export type MutationUpdateProfileArgs = {
  input: UpdateProfileInput;
};

/** The final result recorded for a decision. */
export type Outcome = {
  __typename?: 'Outcome';
  createdAt: Scalars['DateTime']['output'];
  decision: Decision;
  id: Scalars['ID']['output'];
  resolvedAt: Scalars['DateTime']['output'];
  success: Scalars['Boolean']['output'];
  summary: Scalars['String']['output'];
  updatedAt: Scalars['DateTime']['output'];
};

export type OutcomePayload = {
  __typename?: 'OutcomePayload';
  outcome?: Maybe<Outcome>;
  userErrors: Array<UserError>;
};

/** Relay-style pagination metadata for a connection. */
export type PageInfo = {
  __typename?: 'PageInfo';
  endCursor?: Maybe<Scalars['String']['output']>;
  hasNextPage: Scalars['Boolean']['output'];
  hasPreviousPage: Scalars['Boolean']['output'];
  startCursor?: Maybe<Scalars['String']['output']>;
};

/** A forecast attached to a decision, with a probability in [0, 1]. */
export type Prediction = {
  __typename?: 'Prediction';
  createdAt: Scalars['DateTime']['output'];
  decision: Decision;
  id: Scalars['ID']['output'];
  probability: Scalars['Float']['output'];
  resolvesAt?: Maybe<Scalars['DateTime']['output']>;
  statement: Scalars['String']['output'];
  updatedAt: Scalars['DateTime']['output'];
};

export type PredictionPayload = {
  __typename?: 'PredictionPayload';
  prediction?: Maybe<Prediction>;
  userErrors: Array<UserError>;
};

export type Query = {
  __typename?: 'Query';
  /** Calibration report for a team. */
  calibration: CalibrationReport;
  /** A single decision by id. */
  decision?: Maybe<Decision>;
  /** A paginated list of a team's decisions. */
  decisions: DecisionConnection;
  /** Liveness probe; returns a static value when the API is reachable. */
  health: Scalars['String']['output'];
  /** The currently authenticated user. */
  me?: Maybe<User>;
  /** Teams the authenticated user belongs to. */
  myTeams: Array<Team>;
  /** A single team by id. */
  team?: Maybe<Team>;
  /** Decision-quality metrics for a team. */
  teamMetrics: TeamMetrics;
  /** A single user by id. */
  user?: Maybe<User>;
};


export type QueryCalibrationArgs = {
  teamId: Scalars['ID']['input'];
};


export type QueryDecisionArgs = {
  id: Scalars['ID']['input'];
};


export type QueryDecisionsArgs = {
  after?: InputMaybe<Scalars['String']['input']>;
  first?: InputMaybe<Scalars['Int']['input']>;
  teamId: Scalars['ID']['input'];
};


export type QueryTeamArgs = {
  id: Scalars['ID']['input'];
};


export type QueryTeamMetricsArgs = {
  teamId: Scalars['ID']['input'];
};


export type QueryUserArgs = {
  id: Scalars['ID']['input'];
};

export type RecordOutcomeInput = {
  decisionId: Scalars['ID']['input'];
  resolvedAt: Scalars['DateTime']['input'];
  success: Scalars['Boolean']['input'];
  summary: Scalars['String']['input'];
};

export type RegisterInput = {
  email: Scalars['String']['input'];
  name: Scalars['String']['input'];
  password: Scalars['String']['input'];
};

export type RemoveTeamMemberInput = {
  teamId: Scalars['ID']['input'];
  userId: Scalars['ID']['input'];
};

export type RemoveTeamMemberPayload = {
  __typename?: 'RemoveTeamMemberPayload';
  success: Scalars['Boolean']['output'];
  userErrors: Array<UserError>;
};

/** Membership role within a team, ordered from most to least privileged. */
export type Role =
  | 'ADMIN'
  | 'MEMBER'
  | 'VIEWER';

export type Subscription = {
  __typename?: 'Subscription';
  /** Emits a decision whenever one changes within the given team. */
  decisionUpdated: Decision;
};


export type SubscriptionDecisionUpdatedArgs = {
  teamId: Scalars['ID']['input'];
};

/** A team that groups users and scopes their decisions. */
export type Team = {
  __typename?: 'Team';
  createdAt: Scalars['DateTime']['output'];
  id: Scalars['ID']['output'];
  members: Array<Membership>;
  name: Scalars['String']['output'];
  updatedAt: Scalars['DateTime']['output'];
};

/** Headline decision-quality summary for a team. */
export type TeamMetrics = {
  __typename?: 'TeamMetrics';
  /** Mean squared error of resolved forecasts, in [0, 1]; lower is better. */
  brierScore: Scalars['Float']['output'];
  /** Fraction of resolved decisions that succeeded, in [0, 1]. */
  decisionSuccessRate: Scalars['Float']['output'];
  /** Number of resolved forecasts the Brier score covers. */
  forecastCount: Scalars['Int']['output'];
  /** Number of decisions with a recorded outcome. */
  resolvedDecisionCount: Scalars['Int']['output'];
};

export type TransitionDecisionInput = {
  id: Scalars['ID']['input'];
  status: DecisionStatus;
};

export type UpdateAssumptionInput = {
  confidence: Scalars['Float']['input'];
  id: Scalars['ID']['input'];
  statement: Scalars['String']['input'];
};

export type UpdateDecisionInput = {
  description: Scalars['String']['input'];
  id: Scalars['ID']['input'];
  title: Scalars['String']['input'];
};

export type UpdateEvidenceInput = {
  content: Scalars['String']['input'];
  id: Scalars['ID']['input'];
  sourceType: EvidenceSourceType;
  sourceUrl?: InputMaybe<Scalars['String']['input']>;
};

export type UpdatePredictionInput = {
  id: Scalars['ID']['input'];
  probability: Scalars['Float']['input'];
  resolvesAt?: InputMaybe<Scalars['DateTime']['input']>;
  statement: Scalars['String']['input'];
};

export type UpdateProfileInput = {
  name: Scalars['String']['input'];
};

export type UpdateProfilePayload = {
  __typename?: 'UpdateProfilePayload';
  user?: Maybe<User>;
  userErrors: Array<UserError>;
};

/** A user account. */
export type User = {
  __typename?: 'User';
  createdAt: Scalars['DateTime']['output'];
  email: Scalars['String']['output'];
  id: Scalars['ID']['output'];
  name: Scalars['String']['output'];
  updatedAt: Scalars['DateTime']['output'];
};

/** A recoverable, field-level error returned from a mutation. */
export type UserError = {
  __typename?: 'UserError';
  /** A stable, machine-readable error code. */
  code: Scalars['String']['output'];
  /** The input field the error applies to, if any. */
  field?: Maybe<Scalars['String']['output']>;
  /** A human-readable description of the error. */
  message: Scalars['String']['output'];
};

export type TeamMetricsQueryVariables = Exact<{
  teamId: Scalars['ID']['input'];
}>;


export type TeamMetricsQuery = { __typename?: 'Query', teamMetrics: { __typename?: 'TeamMetrics', brierScore: number, forecastCount: number, decisionSuccessRate: number, resolvedDecisionCount: number } };

export type CalibrationQueryVariables = Exact<{
  teamId: Scalars['ID']['input'];
}>;


export type CalibrationQuery = { __typename?: 'Query', calibration: { __typename?: 'CalibrationReport', bins: Array<{ __typename?: 'CalibrationBin', bucket: number, meanPredicted: number, observedFrequency: number, sampleSize: number }> } };

export type AuthPayloadFieldsFragment = { __typename?: 'AuthPayload', accessToken?: string | null, accessExpiresAt?: string | null, user?: { __typename?: 'User', id: string, email: string, name: string } | null, userErrors: Array<{ __typename?: 'UserError', field?: string | null, message: string, code: string }> };

export type LoginMutationVariables = Exact<{
  input: LoginInput;
}>;


export type LoginMutation = { __typename?: 'Mutation', login: { __typename?: 'AuthPayload', accessToken?: string | null, accessExpiresAt?: string | null, user?: { __typename?: 'User', id: string, email: string, name: string } | null, userErrors: Array<{ __typename?: 'UserError', field?: string | null, message: string, code: string }> } };

export type RegisterMutationVariables = Exact<{
  input: RegisterInput;
}>;


export type RegisterMutation = { __typename?: 'Mutation', register: { __typename?: 'AuthPayload', accessToken?: string | null, accessExpiresAt?: string | null, user?: { __typename?: 'User', id: string, email: string, name: string } | null, userErrors: Array<{ __typename?: 'UserError', field?: string | null, message: string, code: string }> } };

export type RefreshMutationVariables = Exact<{ [key: string]: never; }>;


export type RefreshMutation = { __typename?: 'Mutation', refreshToken: { __typename?: 'AuthPayload', accessToken?: string | null, accessExpiresAt?: string | null, user?: { __typename?: 'User', id: string, email: string, name: string } | null, userErrors: Array<{ __typename?: 'UserError', field?: string | null, message: string, code: string }> } };

export type LogoutMutationVariables = Exact<{ [key: string]: never; }>;


export type LogoutMutation = { __typename?: 'Mutation', logout: boolean };

export type MeQueryVariables = Exact<{ [key: string]: never; }>;


export type MeQuery = { __typename?: 'Query', me?: { __typename?: 'User', id: string, email: string, name: string } | null };

export type EvidenceFieldsFragment = { __typename?: 'Evidence', id: string, sourceType: EvidenceSourceType, sourceUrl?: string | null, content: string, createdAt: string };

export type AssumptionFieldsFragment = { __typename?: 'Assumption', id: string, statement: string, confidence: number, createdAt: string, evidence: Array<{ __typename?: 'Evidence', id: string, sourceType: EvidenceSourceType, sourceUrl?: string | null, content: string, createdAt: string }> };

export type PredictionFieldsFragment = { __typename?: 'Prediction', id: string, statement: string, probability: number, resolvesAt?: string | null, createdAt: string };

export type OutcomeFieldsFragment = { __typename?: 'Outcome', id: string, summary: string, success: boolean, resolvedAt: string };

export type DecisionDetailFragment = { __typename?: 'Decision', id: string, title: string, description: string, status: DecisionStatus, decidedAt?: string | null, createdAt: string, updatedAt: string, owner: { __typename?: 'User', id: string, name: string, email: string }, team: { __typename?: 'Team', id: string, name: string }, assumptions: Array<{ __typename?: 'Assumption', id: string, statement: string, confidence: number, createdAt: string, evidence: Array<{ __typename?: 'Evidence', id: string, sourceType: EvidenceSourceType, sourceUrl?: string | null, content: string, createdAt: string }> }>, predictions: Array<{ __typename?: 'Prediction', id: string, statement: string, probability: number, resolvesAt?: string | null, createdAt: string }>, outcome?: { __typename?: 'Outcome', id: string, summary: string, success: boolean, resolvedAt: string } | null };

export type DecisionQueryVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type DecisionQuery = { __typename?: 'Query', decision?: { __typename?: 'Decision', id: string, title: string, description: string, status: DecisionStatus, decidedAt?: string | null, createdAt: string, updatedAt: string, owner: { __typename?: 'User', id: string, name: string, email: string }, team: { __typename?: 'Team', id: string, name: string }, assumptions: Array<{ __typename?: 'Assumption', id: string, statement: string, confidence: number, createdAt: string, evidence: Array<{ __typename?: 'Evidence', id: string, sourceType: EvidenceSourceType, sourceUrl?: string | null, content: string, createdAt: string }> }>, predictions: Array<{ __typename?: 'Prediction', id: string, statement: string, probability: number, resolvesAt?: string | null, createdAt: string }>, outcome?: { __typename?: 'Outcome', id: string, summary: string, success: boolean, resolvedAt: string } | null } | null };

export type UpdateDecisionMutationVariables = Exact<{
  input: UpdateDecisionInput;
}>;


export type UpdateDecisionMutation = { __typename?: 'Mutation', updateDecision: { __typename?: 'DecisionPayload', decision?: { __typename?: 'Decision', id: string, title: string, description: string } | null, userErrors: Array<{ __typename?: 'UserError', field?: string | null, message: string, code: string }> } };

export type TransitionDecisionMutationVariables = Exact<{
  input: TransitionDecisionInput;
}>;


export type TransitionDecisionMutation = { __typename?: 'Mutation', transitionDecision: { __typename?: 'DecisionPayload', decision?: { __typename?: 'Decision', id: string, status: DecisionStatus, decidedAt?: string | null } | null, userErrors: Array<{ __typename?: 'UserError', field?: string | null, message: string, code: string }> } };

export type AddAssumptionMutationVariables = Exact<{
  input: AddAssumptionInput;
}>;


export type AddAssumptionMutation = { __typename?: 'Mutation', addAssumption: { __typename?: 'AssumptionPayload', assumption?: { __typename?: 'Assumption', id: string, statement: string, confidence: number, createdAt: string, evidence: Array<{ __typename?: 'Evidence', id: string, sourceType: EvidenceSourceType, sourceUrl?: string | null, content: string, createdAt: string }> } | null, userErrors: Array<{ __typename?: 'UserError', field?: string | null, message: string, code: string }> } };

export type UpdateAssumptionMutationVariables = Exact<{
  input: UpdateAssumptionInput;
}>;


export type UpdateAssumptionMutation = { __typename?: 'Mutation', updateAssumption: { __typename?: 'AssumptionPayload', assumption?: { __typename?: 'Assumption', id: string, statement: string, confidence: number, createdAt: string, evidence: Array<{ __typename?: 'Evidence', id: string, sourceType: EvidenceSourceType, sourceUrl?: string | null, content: string, createdAt: string }> } | null, userErrors: Array<{ __typename?: 'UserError', field?: string | null, message: string, code: string }> } };

export type RemoveAssumptionMutationVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type RemoveAssumptionMutation = { __typename?: 'Mutation', removeAssumption: { __typename?: 'DeletePayload', success: boolean, userErrors: Array<{ __typename?: 'UserError', field?: string | null, message: string, code: string }> } };

export type AttachEvidenceMutationVariables = Exact<{
  input: AttachEvidenceInput;
}>;


export type AttachEvidenceMutation = { __typename?: 'Mutation', attachEvidence: { __typename?: 'EvidencePayload', evidence?: { __typename?: 'Evidence', id: string, sourceType: EvidenceSourceType, sourceUrl?: string | null, content: string, createdAt: string } | null, userErrors: Array<{ __typename?: 'UserError', field?: string | null, message: string, code: string }> } };

export type UpdateEvidenceMutationVariables = Exact<{
  input: UpdateEvidenceInput;
}>;


export type UpdateEvidenceMutation = { __typename?: 'Mutation', updateEvidence: { __typename?: 'EvidencePayload', evidence?: { __typename?: 'Evidence', id: string, sourceType: EvidenceSourceType, sourceUrl?: string | null, content: string, createdAt: string } | null, userErrors: Array<{ __typename?: 'UserError', field?: string | null, message: string, code: string }> } };

export type RemoveEvidenceMutationVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type RemoveEvidenceMutation = { __typename?: 'Mutation', removeEvidence: { __typename?: 'DeletePayload', success: boolean, userErrors: Array<{ __typename?: 'UserError', field?: string | null, message: string, code: string }> } };

export type CreatePredictionMutationVariables = Exact<{
  input: CreatePredictionInput;
}>;


export type CreatePredictionMutation = { __typename?: 'Mutation', createPrediction: { __typename?: 'PredictionPayload', prediction?: { __typename?: 'Prediction', id: string, statement: string, probability: number, resolvesAt?: string | null, createdAt: string } | null, userErrors: Array<{ __typename?: 'UserError', field?: string | null, message: string, code: string }> } };

export type UpdatePredictionMutationVariables = Exact<{
  input: UpdatePredictionInput;
}>;


export type UpdatePredictionMutation = { __typename?: 'Mutation', updatePrediction: { __typename?: 'PredictionPayload', prediction?: { __typename?: 'Prediction', id: string, statement: string, probability: number, resolvesAt?: string | null, createdAt: string } | null, userErrors: Array<{ __typename?: 'UserError', field?: string | null, message: string, code: string }> } };

export type RecordOutcomeMutationVariables = Exact<{
  input: RecordOutcomeInput;
}>;


export type RecordOutcomeMutation = { __typename?: 'Mutation', recordOutcome: { __typename?: 'OutcomePayload', outcome?: { __typename?: 'Outcome', id: string, summary: string, success: boolean, resolvedAt: string } | null, userErrors: Array<{ __typename?: 'UserError', field?: string | null, message: string, code: string }> } };

export type DecisionListItemFragment = { __typename?: 'Decision', id: string, title: string, description: string, status: DecisionStatus, decidedAt?: string | null, createdAt: string, updatedAt: string, owner: { __typename?: 'User', id: string, name: string } };

export type DecisionsQueryVariables = Exact<{
  teamId: Scalars['ID']['input'];
  first?: InputMaybe<Scalars['Int']['input']>;
  after?: InputMaybe<Scalars['String']['input']>;
}>;


export type DecisionsQuery = { __typename?: 'Query', decisions: { __typename?: 'DecisionConnection', totalCount: number, edges: Array<{ __typename?: 'DecisionEdge', cursor: string, node: { __typename?: 'Decision', id: string, title: string, description: string, status: DecisionStatus, decidedAt?: string | null, createdAt: string, updatedAt: string, owner: { __typename?: 'User', id: string, name: string } } }>, pageInfo: { __typename?: 'PageInfo', hasNextPage: boolean, endCursor?: string | null } } };

export type CreateDecisionMutationVariables = Exact<{
  input: CreateDecisionInput;
}>;


export type CreateDecisionMutation = { __typename?: 'Mutation', createDecision: { __typename?: 'DecisionPayload', decision?: { __typename?: 'Decision', id: string, title: string, description: string, status: DecisionStatus, decidedAt?: string | null, createdAt: string, updatedAt: string, owner: { __typename?: 'User', id: string, name: string } } | null, userErrors: Array<{ __typename?: 'UserError', field?: string | null, message: string, code: string }> } };

export type HealthQueryVariables = Exact<{ [key: string]: never; }>;


export type HealthQuery = { __typename?: 'Query', health: string };

export type TeamDetailQueryVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type TeamDetailQuery = { __typename?: 'Query', team?: { __typename?: 'Team', id: string, name: string, createdAt: string, members: Array<{ __typename?: 'Membership', role: Role, createdAt: string, user: { __typename?: 'User', id: string, name: string, email: string } }> } | null };

export type CreateTeamMutationVariables = Exact<{
  input: CreateTeamInput;
}>;


export type CreateTeamMutation = { __typename?: 'Mutation', createTeam: { __typename?: 'CreateTeamPayload', team?: { __typename?: 'Team', id: string, name: string } | null, userErrors: Array<{ __typename?: 'UserError', field?: string | null, message: string, code: string }> } };

export type AddTeamMemberMutationVariables = Exact<{
  input: AddTeamMemberInput;
}>;


export type AddTeamMemberMutation = { __typename?: 'Mutation', addTeamMember: { __typename?: 'AddTeamMemberPayload', membership?: { __typename?: 'Membership', role: Role, user: { __typename?: 'User', id: string, name: string, email: string } } | null, userErrors: Array<{ __typename?: 'UserError', field?: string | null, message: string, code: string }> } };

export type ChangeMemberRoleMutationVariables = Exact<{
  input: ChangeMemberRoleInput;
}>;


export type ChangeMemberRoleMutation = { __typename?: 'Mutation', changeMemberRole: { __typename?: 'ChangeMemberRolePayload', membership?: { __typename?: 'Membership', role: Role, user: { __typename?: 'User', id: string, name: string, email: string } } | null, userErrors: Array<{ __typename?: 'UserError', field?: string | null, message: string, code: string }> } };

export type RemoveTeamMemberMutationVariables = Exact<{
  input: RemoveTeamMemberInput;
}>;


export type RemoveTeamMemberMutation = { __typename?: 'Mutation', removeTeamMember: { __typename?: 'RemoveTeamMemberPayload', success: boolean, userErrors: Array<{ __typename?: 'UserError', field?: string | null, message: string, code: string }> } };

export type MyTeamsQueryVariables = Exact<{ [key: string]: never; }>;


export type MyTeamsQuery = { __typename?: 'Query', myTeams: Array<{ __typename?: 'Team', id: string, name: string, createdAt: string }> };

export const AuthPayloadFieldsFragmentDoc = gql`
    fragment AuthPayloadFields on AuthPayload {
  user {
    id
    email
    name
  }
  accessToken
  accessExpiresAt
  userErrors {
    field
    message
    code
  }
}
    `;
export const EvidenceFieldsFragmentDoc = gql`
    fragment EvidenceFields on Evidence {
  id
  sourceType
  sourceUrl
  content
  createdAt
}
    `;
export const AssumptionFieldsFragmentDoc = gql`
    fragment AssumptionFields on Assumption {
  id
  statement
  confidence
  createdAt
  evidence {
    ...EvidenceFields
  }
}
    ${EvidenceFieldsFragmentDoc}`;
export const PredictionFieldsFragmentDoc = gql`
    fragment PredictionFields on Prediction {
  id
  statement
  probability
  resolvesAt
  createdAt
}
    `;
export const OutcomeFieldsFragmentDoc = gql`
    fragment OutcomeFields on Outcome {
  id
  summary
  success
  resolvedAt
}
    `;
export const DecisionDetailFragmentDoc = gql`
    fragment DecisionDetail on Decision {
  id
  title
  description
  status
  decidedAt
  createdAt
  updatedAt
  owner {
    id
    name
    email
  }
  team {
    id
    name
  }
  assumptions {
    ...AssumptionFields
  }
  predictions {
    ...PredictionFields
  }
  outcome {
    ...OutcomeFields
  }
}
    ${AssumptionFieldsFragmentDoc}
${PredictionFieldsFragmentDoc}
${OutcomeFieldsFragmentDoc}`;
export const DecisionListItemFragmentDoc = gql`
    fragment DecisionListItem on Decision {
  id
  title
  description
  status
  decidedAt
  createdAt
  updatedAt
  owner {
    id
    name
  }
}
    `;
export const TeamMetricsDocument = gql`
    query TeamMetrics($teamId: ID!) {
  teamMetrics(teamId: $teamId) {
    brierScore
    forecastCount
    decisionSuccessRate
    resolvedDecisionCount
  }
}
    `;

/**
 * __useTeamMetricsQuery__
 *
 * To run a query within a React component, call `useTeamMetricsQuery` and pass it any options that fit your needs.
 * When your component renders, `useTeamMetricsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useTeamMetricsQuery({
 *   variables: {
 *      teamId: // value for 'teamId'
 *   },
 * });
 */
export function useTeamMetricsQuery(baseOptions: Apollo.QueryHookOptions<TeamMetricsQuery, TeamMetricsQueryVariables> & ({ variables: TeamMetricsQueryVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<TeamMetricsQuery, TeamMetricsQueryVariables>(TeamMetricsDocument, options);
      }
export function useTeamMetricsLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<TeamMetricsQuery, TeamMetricsQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<TeamMetricsQuery, TeamMetricsQueryVariables>(TeamMetricsDocument, options);
        }
// @ts-ignore
export function useTeamMetricsSuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<TeamMetricsQuery, TeamMetricsQueryVariables>): Apollo.UseSuspenseQueryResult<TeamMetricsQuery, TeamMetricsQueryVariables>;
export function useTeamMetricsSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<TeamMetricsQuery, TeamMetricsQueryVariables>): Apollo.UseSuspenseQueryResult<TeamMetricsQuery | undefined, TeamMetricsQueryVariables>;
export function useTeamMetricsSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<TeamMetricsQuery, TeamMetricsQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<TeamMetricsQuery, TeamMetricsQueryVariables>(TeamMetricsDocument, options);
        }
export type TeamMetricsQueryHookResult = ReturnType<typeof useTeamMetricsQuery>;
export type TeamMetricsLazyQueryHookResult = ReturnType<typeof useTeamMetricsLazyQuery>;
export type TeamMetricsSuspenseQueryHookResult = ReturnType<typeof useTeamMetricsSuspenseQuery>;
export type TeamMetricsQueryResult = Apollo.QueryResult<TeamMetricsQuery, TeamMetricsQueryVariables>;
export const CalibrationDocument = gql`
    query Calibration($teamId: ID!) {
  calibration(teamId: $teamId) {
    bins {
      bucket
      meanPredicted
      observedFrequency
      sampleSize
    }
  }
}
    `;

/**
 * __useCalibrationQuery__
 *
 * To run a query within a React component, call `useCalibrationQuery` and pass it any options that fit your needs.
 * When your component renders, `useCalibrationQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useCalibrationQuery({
 *   variables: {
 *      teamId: // value for 'teamId'
 *   },
 * });
 */
export function useCalibrationQuery(baseOptions: Apollo.QueryHookOptions<CalibrationQuery, CalibrationQueryVariables> & ({ variables: CalibrationQueryVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<CalibrationQuery, CalibrationQueryVariables>(CalibrationDocument, options);
      }
export function useCalibrationLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<CalibrationQuery, CalibrationQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<CalibrationQuery, CalibrationQueryVariables>(CalibrationDocument, options);
        }
// @ts-ignore
export function useCalibrationSuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<CalibrationQuery, CalibrationQueryVariables>): Apollo.UseSuspenseQueryResult<CalibrationQuery, CalibrationQueryVariables>;
export function useCalibrationSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<CalibrationQuery, CalibrationQueryVariables>): Apollo.UseSuspenseQueryResult<CalibrationQuery | undefined, CalibrationQueryVariables>;
export function useCalibrationSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<CalibrationQuery, CalibrationQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<CalibrationQuery, CalibrationQueryVariables>(CalibrationDocument, options);
        }
export type CalibrationQueryHookResult = ReturnType<typeof useCalibrationQuery>;
export type CalibrationLazyQueryHookResult = ReturnType<typeof useCalibrationLazyQuery>;
export type CalibrationSuspenseQueryHookResult = ReturnType<typeof useCalibrationSuspenseQuery>;
export type CalibrationQueryResult = Apollo.QueryResult<CalibrationQuery, CalibrationQueryVariables>;
export const LoginDocument = gql`
    mutation Login($input: LoginInput!) {
  login(input: $input) {
    ...AuthPayloadFields
  }
}
    ${AuthPayloadFieldsFragmentDoc}`;
export type LoginMutationFn = Apollo.MutationFunction<LoginMutation, LoginMutationVariables>;

/**
 * __useLoginMutation__
 *
 * To run a mutation, you first call `useLoginMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useLoginMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [loginMutation, { data, loading, error }] = useLoginMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useLoginMutation(baseOptions?: Apollo.MutationHookOptions<LoginMutation, LoginMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<LoginMutation, LoginMutationVariables>(LoginDocument, options);
      }
export type LoginMutationHookResult = ReturnType<typeof useLoginMutation>;
export type LoginMutationResult = Apollo.MutationResult<LoginMutation>;
export type LoginMutationOptions = Apollo.BaseMutationOptions<LoginMutation, LoginMutationVariables>;
export const RegisterDocument = gql`
    mutation Register($input: RegisterInput!) {
  register(input: $input) {
    ...AuthPayloadFields
  }
}
    ${AuthPayloadFieldsFragmentDoc}`;
export type RegisterMutationFn = Apollo.MutationFunction<RegisterMutation, RegisterMutationVariables>;

/**
 * __useRegisterMutation__
 *
 * To run a mutation, you first call `useRegisterMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useRegisterMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [registerMutation, { data, loading, error }] = useRegisterMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useRegisterMutation(baseOptions?: Apollo.MutationHookOptions<RegisterMutation, RegisterMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<RegisterMutation, RegisterMutationVariables>(RegisterDocument, options);
      }
export type RegisterMutationHookResult = ReturnType<typeof useRegisterMutation>;
export type RegisterMutationResult = Apollo.MutationResult<RegisterMutation>;
export type RegisterMutationOptions = Apollo.BaseMutationOptions<RegisterMutation, RegisterMutationVariables>;
export const RefreshDocument = gql`
    mutation Refresh {
  refreshToken {
    ...AuthPayloadFields
  }
}
    ${AuthPayloadFieldsFragmentDoc}`;
export type RefreshMutationFn = Apollo.MutationFunction<RefreshMutation, RefreshMutationVariables>;

/**
 * __useRefreshMutation__
 *
 * To run a mutation, you first call `useRefreshMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useRefreshMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [refreshMutation, { data, loading, error }] = useRefreshMutation({
 *   variables: {
 *   },
 * });
 */
export function useRefreshMutation(baseOptions?: Apollo.MutationHookOptions<RefreshMutation, RefreshMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<RefreshMutation, RefreshMutationVariables>(RefreshDocument, options);
      }
export type RefreshMutationHookResult = ReturnType<typeof useRefreshMutation>;
export type RefreshMutationResult = Apollo.MutationResult<RefreshMutation>;
export type RefreshMutationOptions = Apollo.BaseMutationOptions<RefreshMutation, RefreshMutationVariables>;
export const LogoutDocument = gql`
    mutation Logout {
  logout
}
    `;
export type LogoutMutationFn = Apollo.MutationFunction<LogoutMutation, LogoutMutationVariables>;

/**
 * __useLogoutMutation__
 *
 * To run a mutation, you first call `useLogoutMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useLogoutMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [logoutMutation, { data, loading, error }] = useLogoutMutation({
 *   variables: {
 *   },
 * });
 */
export function useLogoutMutation(baseOptions?: Apollo.MutationHookOptions<LogoutMutation, LogoutMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<LogoutMutation, LogoutMutationVariables>(LogoutDocument, options);
      }
export type LogoutMutationHookResult = ReturnType<typeof useLogoutMutation>;
export type LogoutMutationResult = Apollo.MutationResult<LogoutMutation>;
export type LogoutMutationOptions = Apollo.BaseMutationOptions<LogoutMutation, LogoutMutationVariables>;
export const MeDocument = gql`
    query Me {
  me {
    id
    email
    name
  }
}
    `;

/**
 * __useMeQuery__
 *
 * To run a query within a React component, call `useMeQuery` and pass it any options that fit your needs.
 * When your component renders, `useMeQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useMeQuery({
 *   variables: {
 *   },
 * });
 */
export function useMeQuery(baseOptions?: Apollo.QueryHookOptions<MeQuery, MeQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<MeQuery, MeQueryVariables>(MeDocument, options);
      }
export function useMeLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<MeQuery, MeQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<MeQuery, MeQueryVariables>(MeDocument, options);
        }
// @ts-ignore
export function useMeSuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<MeQuery, MeQueryVariables>): Apollo.UseSuspenseQueryResult<MeQuery, MeQueryVariables>;
export function useMeSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<MeQuery, MeQueryVariables>): Apollo.UseSuspenseQueryResult<MeQuery | undefined, MeQueryVariables>;
export function useMeSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<MeQuery, MeQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<MeQuery, MeQueryVariables>(MeDocument, options);
        }
export type MeQueryHookResult = ReturnType<typeof useMeQuery>;
export type MeLazyQueryHookResult = ReturnType<typeof useMeLazyQuery>;
export type MeSuspenseQueryHookResult = ReturnType<typeof useMeSuspenseQuery>;
export type MeQueryResult = Apollo.QueryResult<MeQuery, MeQueryVariables>;
export const DecisionDocument = gql`
    query Decision($id: ID!) {
  decision(id: $id) {
    ...DecisionDetail
  }
}
    ${DecisionDetailFragmentDoc}`;

/**
 * __useDecisionQuery__
 *
 * To run a query within a React component, call `useDecisionQuery` and pass it any options that fit your needs.
 * When your component renders, `useDecisionQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useDecisionQuery({
 *   variables: {
 *      id: // value for 'id'
 *   },
 * });
 */
export function useDecisionQuery(baseOptions: Apollo.QueryHookOptions<DecisionQuery, DecisionQueryVariables> & ({ variables: DecisionQueryVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<DecisionQuery, DecisionQueryVariables>(DecisionDocument, options);
      }
export function useDecisionLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<DecisionQuery, DecisionQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<DecisionQuery, DecisionQueryVariables>(DecisionDocument, options);
        }
// @ts-ignore
export function useDecisionSuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<DecisionQuery, DecisionQueryVariables>): Apollo.UseSuspenseQueryResult<DecisionQuery, DecisionQueryVariables>;
export function useDecisionSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<DecisionQuery, DecisionQueryVariables>): Apollo.UseSuspenseQueryResult<DecisionQuery | undefined, DecisionQueryVariables>;
export function useDecisionSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<DecisionQuery, DecisionQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<DecisionQuery, DecisionQueryVariables>(DecisionDocument, options);
        }
export type DecisionQueryHookResult = ReturnType<typeof useDecisionQuery>;
export type DecisionLazyQueryHookResult = ReturnType<typeof useDecisionLazyQuery>;
export type DecisionSuspenseQueryHookResult = ReturnType<typeof useDecisionSuspenseQuery>;
export type DecisionQueryResult = Apollo.QueryResult<DecisionQuery, DecisionQueryVariables>;
export const UpdateDecisionDocument = gql`
    mutation UpdateDecision($input: UpdateDecisionInput!) {
  updateDecision(input: $input) {
    decision {
      id
      title
      description
    }
    userErrors {
      field
      message
      code
    }
  }
}
    `;
export type UpdateDecisionMutationFn = Apollo.MutationFunction<UpdateDecisionMutation, UpdateDecisionMutationVariables>;

/**
 * __useUpdateDecisionMutation__
 *
 * To run a mutation, you first call `useUpdateDecisionMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useUpdateDecisionMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [updateDecisionMutation, { data, loading, error }] = useUpdateDecisionMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useUpdateDecisionMutation(baseOptions?: Apollo.MutationHookOptions<UpdateDecisionMutation, UpdateDecisionMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<UpdateDecisionMutation, UpdateDecisionMutationVariables>(UpdateDecisionDocument, options);
      }
export type UpdateDecisionMutationHookResult = ReturnType<typeof useUpdateDecisionMutation>;
export type UpdateDecisionMutationResult = Apollo.MutationResult<UpdateDecisionMutation>;
export type UpdateDecisionMutationOptions = Apollo.BaseMutationOptions<UpdateDecisionMutation, UpdateDecisionMutationVariables>;
export const TransitionDecisionDocument = gql`
    mutation TransitionDecision($input: TransitionDecisionInput!) {
  transitionDecision(input: $input) {
    decision {
      id
      status
      decidedAt
    }
    userErrors {
      field
      message
      code
    }
  }
}
    `;
export type TransitionDecisionMutationFn = Apollo.MutationFunction<TransitionDecisionMutation, TransitionDecisionMutationVariables>;

/**
 * __useTransitionDecisionMutation__
 *
 * To run a mutation, you first call `useTransitionDecisionMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useTransitionDecisionMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [transitionDecisionMutation, { data, loading, error }] = useTransitionDecisionMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useTransitionDecisionMutation(baseOptions?: Apollo.MutationHookOptions<TransitionDecisionMutation, TransitionDecisionMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<TransitionDecisionMutation, TransitionDecisionMutationVariables>(TransitionDecisionDocument, options);
      }
export type TransitionDecisionMutationHookResult = ReturnType<typeof useTransitionDecisionMutation>;
export type TransitionDecisionMutationResult = Apollo.MutationResult<TransitionDecisionMutation>;
export type TransitionDecisionMutationOptions = Apollo.BaseMutationOptions<TransitionDecisionMutation, TransitionDecisionMutationVariables>;
export const AddAssumptionDocument = gql`
    mutation AddAssumption($input: AddAssumptionInput!) {
  addAssumption(input: $input) {
    assumption {
      ...AssumptionFields
    }
    userErrors {
      field
      message
      code
    }
  }
}
    ${AssumptionFieldsFragmentDoc}`;
export type AddAssumptionMutationFn = Apollo.MutationFunction<AddAssumptionMutation, AddAssumptionMutationVariables>;

/**
 * __useAddAssumptionMutation__
 *
 * To run a mutation, you first call `useAddAssumptionMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useAddAssumptionMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [addAssumptionMutation, { data, loading, error }] = useAddAssumptionMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useAddAssumptionMutation(baseOptions?: Apollo.MutationHookOptions<AddAssumptionMutation, AddAssumptionMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<AddAssumptionMutation, AddAssumptionMutationVariables>(AddAssumptionDocument, options);
      }
export type AddAssumptionMutationHookResult = ReturnType<typeof useAddAssumptionMutation>;
export type AddAssumptionMutationResult = Apollo.MutationResult<AddAssumptionMutation>;
export type AddAssumptionMutationOptions = Apollo.BaseMutationOptions<AddAssumptionMutation, AddAssumptionMutationVariables>;
export const UpdateAssumptionDocument = gql`
    mutation UpdateAssumption($input: UpdateAssumptionInput!) {
  updateAssumption(input: $input) {
    assumption {
      ...AssumptionFields
    }
    userErrors {
      field
      message
      code
    }
  }
}
    ${AssumptionFieldsFragmentDoc}`;
export type UpdateAssumptionMutationFn = Apollo.MutationFunction<UpdateAssumptionMutation, UpdateAssumptionMutationVariables>;

/**
 * __useUpdateAssumptionMutation__
 *
 * To run a mutation, you first call `useUpdateAssumptionMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useUpdateAssumptionMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [updateAssumptionMutation, { data, loading, error }] = useUpdateAssumptionMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useUpdateAssumptionMutation(baseOptions?: Apollo.MutationHookOptions<UpdateAssumptionMutation, UpdateAssumptionMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<UpdateAssumptionMutation, UpdateAssumptionMutationVariables>(UpdateAssumptionDocument, options);
      }
export type UpdateAssumptionMutationHookResult = ReturnType<typeof useUpdateAssumptionMutation>;
export type UpdateAssumptionMutationResult = Apollo.MutationResult<UpdateAssumptionMutation>;
export type UpdateAssumptionMutationOptions = Apollo.BaseMutationOptions<UpdateAssumptionMutation, UpdateAssumptionMutationVariables>;
export const RemoveAssumptionDocument = gql`
    mutation RemoveAssumption($id: ID!) {
  removeAssumption(id: $id) {
    success
    userErrors {
      field
      message
      code
    }
  }
}
    `;
export type RemoveAssumptionMutationFn = Apollo.MutationFunction<RemoveAssumptionMutation, RemoveAssumptionMutationVariables>;

/**
 * __useRemoveAssumptionMutation__
 *
 * To run a mutation, you first call `useRemoveAssumptionMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useRemoveAssumptionMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [removeAssumptionMutation, { data, loading, error }] = useRemoveAssumptionMutation({
 *   variables: {
 *      id: // value for 'id'
 *   },
 * });
 */
export function useRemoveAssumptionMutation(baseOptions?: Apollo.MutationHookOptions<RemoveAssumptionMutation, RemoveAssumptionMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<RemoveAssumptionMutation, RemoveAssumptionMutationVariables>(RemoveAssumptionDocument, options);
      }
export type RemoveAssumptionMutationHookResult = ReturnType<typeof useRemoveAssumptionMutation>;
export type RemoveAssumptionMutationResult = Apollo.MutationResult<RemoveAssumptionMutation>;
export type RemoveAssumptionMutationOptions = Apollo.BaseMutationOptions<RemoveAssumptionMutation, RemoveAssumptionMutationVariables>;
export const AttachEvidenceDocument = gql`
    mutation AttachEvidence($input: AttachEvidenceInput!) {
  attachEvidence(input: $input) {
    evidence {
      ...EvidenceFields
    }
    userErrors {
      field
      message
      code
    }
  }
}
    ${EvidenceFieldsFragmentDoc}`;
export type AttachEvidenceMutationFn = Apollo.MutationFunction<AttachEvidenceMutation, AttachEvidenceMutationVariables>;

/**
 * __useAttachEvidenceMutation__
 *
 * To run a mutation, you first call `useAttachEvidenceMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useAttachEvidenceMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [attachEvidenceMutation, { data, loading, error }] = useAttachEvidenceMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useAttachEvidenceMutation(baseOptions?: Apollo.MutationHookOptions<AttachEvidenceMutation, AttachEvidenceMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<AttachEvidenceMutation, AttachEvidenceMutationVariables>(AttachEvidenceDocument, options);
      }
export type AttachEvidenceMutationHookResult = ReturnType<typeof useAttachEvidenceMutation>;
export type AttachEvidenceMutationResult = Apollo.MutationResult<AttachEvidenceMutation>;
export type AttachEvidenceMutationOptions = Apollo.BaseMutationOptions<AttachEvidenceMutation, AttachEvidenceMutationVariables>;
export const UpdateEvidenceDocument = gql`
    mutation UpdateEvidence($input: UpdateEvidenceInput!) {
  updateEvidence(input: $input) {
    evidence {
      ...EvidenceFields
    }
    userErrors {
      field
      message
      code
    }
  }
}
    ${EvidenceFieldsFragmentDoc}`;
export type UpdateEvidenceMutationFn = Apollo.MutationFunction<UpdateEvidenceMutation, UpdateEvidenceMutationVariables>;

/**
 * __useUpdateEvidenceMutation__
 *
 * To run a mutation, you first call `useUpdateEvidenceMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useUpdateEvidenceMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [updateEvidenceMutation, { data, loading, error }] = useUpdateEvidenceMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useUpdateEvidenceMutation(baseOptions?: Apollo.MutationHookOptions<UpdateEvidenceMutation, UpdateEvidenceMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<UpdateEvidenceMutation, UpdateEvidenceMutationVariables>(UpdateEvidenceDocument, options);
      }
export type UpdateEvidenceMutationHookResult = ReturnType<typeof useUpdateEvidenceMutation>;
export type UpdateEvidenceMutationResult = Apollo.MutationResult<UpdateEvidenceMutation>;
export type UpdateEvidenceMutationOptions = Apollo.BaseMutationOptions<UpdateEvidenceMutation, UpdateEvidenceMutationVariables>;
export const RemoveEvidenceDocument = gql`
    mutation RemoveEvidence($id: ID!) {
  removeEvidence(id: $id) {
    success
    userErrors {
      field
      message
      code
    }
  }
}
    `;
export type RemoveEvidenceMutationFn = Apollo.MutationFunction<RemoveEvidenceMutation, RemoveEvidenceMutationVariables>;

/**
 * __useRemoveEvidenceMutation__
 *
 * To run a mutation, you first call `useRemoveEvidenceMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useRemoveEvidenceMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [removeEvidenceMutation, { data, loading, error }] = useRemoveEvidenceMutation({
 *   variables: {
 *      id: // value for 'id'
 *   },
 * });
 */
export function useRemoveEvidenceMutation(baseOptions?: Apollo.MutationHookOptions<RemoveEvidenceMutation, RemoveEvidenceMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<RemoveEvidenceMutation, RemoveEvidenceMutationVariables>(RemoveEvidenceDocument, options);
      }
export type RemoveEvidenceMutationHookResult = ReturnType<typeof useRemoveEvidenceMutation>;
export type RemoveEvidenceMutationResult = Apollo.MutationResult<RemoveEvidenceMutation>;
export type RemoveEvidenceMutationOptions = Apollo.BaseMutationOptions<RemoveEvidenceMutation, RemoveEvidenceMutationVariables>;
export const CreatePredictionDocument = gql`
    mutation CreatePrediction($input: CreatePredictionInput!) {
  createPrediction(input: $input) {
    prediction {
      ...PredictionFields
    }
    userErrors {
      field
      message
      code
    }
  }
}
    ${PredictionFieldsFragmentDoc}`;
export type CreatePredictionMutationFn = Apollo.MutationFunction<CreatePredictionMutation, CreatePredictionMutationVariables>;

/**
 * __useCreatePredictionMutation__
 *
 * To run a mutation, you first call `useCreatePredictionMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreatePredictionMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createPredictionMutation, { data, loading, error }] = useCreatePredictionMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useCreatePredictionMutation(baseOptions?: Apollo.MutationHookOptions<CreatePredictionMutation, CreatePredictionMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<CreatePredictionMutation, CreatePredictionMutationVariables>(CreatePredictionDocument, options);
      }
export type CreatePredictionMutationHookResult = ReturnType<typeof useCreatePredictionMutation>;
export type CreatePredictionMutationResult = Apollo.MutationResult<CreatePredictionMutation>;
export type CreatePredictionMutationOptions = Apollo.BaseMutationOptions<CreatePredictionMutation, CreatePredictionMutationVariables>;
export const UpdatePredictionDocument = gql`
    mutation UpdatePrediction($input: UpdatePredictionInput!) {
  updatePrediction(input: $input) {
    prediction {
      ...PredictionFields
    }
    userErrors {
      field
      message
      code
    }
  }
}
    ${PredictionFieldsFragmentDoc}`;
export type UpdatePredictionMutationFn = Apollo.MutationFunction<UpdatePredictionMutation, UpdatePredictionMutationVariables>;

/**
 * __useUpdatePredictionMutation__
 *
 * To run a mutation, you first call `useUpdatePredictionMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useUpdatePredictionMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [updatePredictionMutation, { data, loading, error }] = useUpdatePredictionMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useUpdatePredictionMutation(baseOptions?: Apollo.MutationHookOptions<UpdatePredictionMutation, UpdatePredictionMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<UpdatePredictionMutation, UpdatePredictionMutationVariables>(UpdatePredictionDocument, options);
      }
export type UpdatePredictionMutationHookResult = ReturnType<typeof useUpdatePredictionMutation>;
export type UpdatePredictionMutationResult = Apollo.MutationResult<UpdatePredictionMutation>;
export type UpdatePredictionMutationOptions = Apollo.BaseMutationOptions<UpdatePredictionMutation, UpdatePredictionMutationVariables>;
export const RecordOutcomeDocument = gql`
    mutation RecordOutcome($input: RecordOutcomeInput!) {
  recordOutcome(input: $input) {
    outcome {
      ...OutcomeFields
    }
    userErrors {
      field
      message
      code
    }
  }
}
    ${OutcomeFieldsFragmentDoc}`;
export type RecordOutcomeMutationFn = Apollo.MutationFunction<RecordOutcomeMutation, RecordOutcomeMutationVariables>;

/**
 * __useRecordOutcomeMutation__
 *
 * To run a mutation, you first call `useRecordOutcomeMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useRecordOutcomeMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [recordOutcomeMutation, { data, loading, error }] = useRecordOutcomeMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useRecordOutcomeMutation(baseOptions?: Apollo.MutationHookOptions<RecordOutcomeMutation, RecordOutcomeMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<RecordOutcomeMutation, RecordOutcomeMutationVariables>(RecordOutcomeDocument, options);
      }
export type RecordOutcomeMutationHookResult = ReturnType<typeof useRecordOutcomeMutation>;
export type RecordOutcomeMutationResult = Apollo.MutationResult<RecordOutcomeMutation>;
export type RecordOutcomeMutationOptions = Apollo.BaseMutationOptions<RecordOutcomeMutation, RecordOutcomeMutationVariables>;
export const DecisionsDocument = gql`
    query Decisions($teamId: ID!, $first: Int, $after: String) {
  decisions(teamId: $teamId, first: $first, after: $after) {
    edges {
      node {
        ...DecisionListItem
      }
      cursor
    }
    pageInfo {
      hasNextPage
      endCursor
    }
    totalCount
  }
}
    ${DecisionListItemFragmentDoc}`;

/**
 * __useDecisionsQuery__
 *
 * To run a query within a React component, call `useDecisionsQuery` and pass it any options that fit your needs.
 * When your component renders, `useDecisionsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useDecisionsQuery({
 *   variables: {
 *      teamId: // value for 'teamId'
 *      first: // value for 'first'
 *      after: // value for 'after'
 *   },
 * });
 */
export function useDecisionsQuery(baseOptions: Apollo.QueryHookOptions<DecisionsQuery, DecisionsQueryVariables> & ({ variables: DecisionsQueryVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<DecisionsQuery, DecisionsQueryVariables>(DecisionsDocument, options);
      }
export function useDecisionsLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<DecisionsQuery, DecisionsQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<DecisionsQuery, DecisionsQueryVariables>(DecisionsDocument, options);
        }
// @ts-ignore
export function useDecisionsSuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<DecisionsQuery, DecisionsQueryVariables>): Apollo.UseSuspenseQueryResult<DecisionsQuery, DecisionsQueryVariables>;
export function useDecisionsSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<DecisionsQuery, DecisionsQueryVariables>): Apollo.UseSuspenseQueryResult<DecisionsQuery | undefined, DecisionsQueryVariables>;
export function useDecisionsSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<DecisionsQuery, DecisionsQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<DecisionsQuery, DecisionsQueryVariables>(DecisionsDocument, options);
        }
export type DecisionsQueryHookResult = ReturnType<typeof useDecisionsQuery>;
export type DecisionsLazyQueryHookResult = ReturnType<typeof useDecisionsLazyQuery>;
export type DecisionsSuspenseQueryHookResult = ReturnType<typeof useDecisionsSuspenseQuery>;
export type DecisionsQueryResult = Apollo.QueryResult<DecisionsQuery, DecisionsQueryVariables>;
export const CreateDecisionDocument = gql`
    mutation CreateDecision($input: CreateDecisionInput!) {
  createDecision(input: $input) {
    decision {
      ...DecisionListItem
    }
    userErrors {
      field
      message
      code
    }
  }
}
    ${DecisionListItemFragmentDoc}`;
export type CreateDecisionMutationFn = Apollo.MutationFunction<CreateDecisionMutation, CreateDecisionMutationVariables>;

/**
 * __useCreateDecisionMutation__
 *
 * To run a mutation, you first call `useCreateDecisionMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreateDecisionMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createDecisionMutation, { data, loading, error }] = useCreateDecisionMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useCreateDecisionMutation(baseOptions?: Apollo.MutationHookOptions<CreateDecisionMutation, CreateDecisionMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<CreateDecisionMutation, CreateDecisionMutationVariables>(CreateDecisionDocument, options);
      }
export type CreateDecisionMutationHookResult = ReturnType<typeof useCreateDecisionMutation>;
export type CreateDecisionMutationResult = Apollo.MutationResult<CreateDecisionMutation>;
export type CreateDecisionMutationOptions = Apollo.BaseMutationOptions<CreateDecisionMutation, CreateDecisionMutationVariables>;
export const HealthDocument = gql`
    query Health {
  health
}
    `;

/**
 * __useHealthQuery__
 *
 * To run a query within a React component, call `useHealthQuery` and pass it any options that fit your needs.
 * When your component renders, `useHealthQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useHealthQuery({
 *   variables: {
 *   },
 * });
 */
export function useHealthQuery(baseOptions?: Apollo.QueryHookOptions<HealthQuery, HealthQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<HealthQuery, HealthQueryVariables>(HealthDocument, options);
      }
export function useHealthLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<HealthQuery, HealthQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<HealthQuery, HealthQueryVariables>(HealthDocument, options);
        }
// @ts-ignore
export function useHealthSuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<HealthQuery, HealthQueryVariables>): Apollo.UseSuspenseQueryResult<HealthQuery, HealthQueryVariables>;
export function useHealthSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<HealthQuery, HealthQueryVariables>): Apollo.UseSuspenseQueryResult<HealthQuery | undefined, HealthQueryVariables>;
export function useHealthSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<HealthQuery, HealthQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<HealthQuery, HealthQueryVariables>(HealthDocument, options);
        }
export type HealthQueryHookResult = ReturnType<typeof useHealthQuery>;
export type HealthLazyQueryHookResult = ReturnType<typeof useHealthLazyQuery>;
export type HealthSuspenseQueryHookResult = ReturnType<typeof useHealthSuspenseQuery>;
export type HealthQueryResult = Apollo.QueryResult<HealthQuery, HealthQueryVariables>;
export const TeamDetailDocument = gql`
    query TeamDetail($id: ID!) {
  team(id: $id) {
    id
    name
    createdAt
    members {
      user {
        id
        name
        email
      }
      role
      createdAt
    }
  }
}
    `;

/**
 * __useTeamDetailQuery__
 *
 * To run a query within a React component, call `useTeamDetailQuery` and pass it any options that fit your needs.
 * When your component renders, `useTeamDetailQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useTeamDetailQuery({
 *   variables: {
 *      id: // value for 'id'
 *   },
 * });
 */
export function useTeamDetailQuery(baseOptions: Apollo.QueryHookOptions<TeamDetailQuery, TeamDetailQueryVariables> & ({ variables: TeamDetailQueryVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<TeamDetailQuery, TeamDetailQueryVariables>(TeamDetailDocument, options);
      }
export function useTeamDetailLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<TeamDetailQuery, TeamDetailQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<TeamDetailQuery, TeamDetailQueryVariables>(TeamDetailDocument, options);
        }
// @ts-ignore
export function useTeamDetailSuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<TeamDetailQuery, TeamDetailQueryVariables>): Apollo.UseSuspenseQueryResult<TeamDetailQuery, TeamDetailQueryVariables>;
export function useTeamDetailSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<TeamDetailQuery, TeamDetailQueryVariables>): Apollo.UseSuspenseQueryResult<TeamDetailQuery | undefined, TeamDetailQueryVariables>;
export function useTeamDetailSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<TeamDetailQuery, TeamDetailQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<TeamDetailQuery, TeamDetailQueryVariables>(TeamDetailDocument, options);
        }
export type TeamDetailQueryHookResult = ReturnType<typeof useTeamDetailQuery>;
export type TeamDetailLazyQueryHookResult = ReturnType<typeof useTeamDetailLazyQuery>;
export type TeamDetailSuspenseQueryHookResult = ReturnType<typeof useTeamDetailSuspenseQuery>;
export type TeamDetailQueryResult = Apollo.QueryResult<TeamDetailQuery, TeamDetailQueryVariables>;
export const CreateTeamDocument = gql`
    mutation CreateTeam($input: CreateTeamInput!) {
  createTeam(input: $input) {
    team {
      id
      name
    }
    userErrors {
      field
      message
      code
    }
  }
}
    `;
export type CreateTeamMutationFn = Apollo.MutationFunction<CreateTeamMutation, CreateTeamMutationVariables>;

/**
 * __useCreateTeamMutation__
 *
 * To run a mutation, you first call `useCreateTeamMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreateTeamMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createTeamMutation, { data, loading, error }] = useCreateTeamMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useCreateTeamMutation(baseOptions?: Apollo.MutationHookOptions<CreateTeamMutation, CreateTeamMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<CreateTeamMutation, CreateTeamMutationVariables>(CreateTeamDocument, options);
      }
export type CreateTeamMutationHookResult = ReturnType<typeof useCreateTeamMutation>;
export type CreateTeamMutationResult = Apollo.MutationResult<CreateTeamMutation>;
export type CreateTeamMutationOptions = Apollo.BaseMutationOptions<CreateTeamMutation, CreateTeamMutationVariables>;
export const AddTeamMemberDocument = gql`
    mutation AddTeamMember($input: AddTeamMemberInput!) {
  addTeamMember(input: $input) {
    membership {
      user {
        id
        name
        email
      }
      role
    }
    userErrors {
      field
      message
      code
    }
  }
}
    `;
export type AddTeamMemberMutationFn = Apollo.MutationFunction<AddTeamMemberMutation, AddTeamMemberMutationVariables>;

/**
 * __useAddTeamMemberMutation__
 *
 * To run a mutation, you first call `useAddTeamMemberMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useAddTeamMemberMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [addTeamMemberMutation, { data, loading, error }] = useAddTeamMemberMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useAddTeamMemberMutation(baseOptions?: Apollo.MutationHookOptions<AddTeamMemberMutation, AddTeamMemberMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<AddTeamMemberMutation, AddTeamMemberMutationVariables>(AddTeamMemberDocument, options);
      }
export type AddTeamMemberMutationHookResult = ReturnType<typeof useAddTeamMemberMutation>;
export type AddTeamMemberMutationResult = Apollo.MutationResult<AddTeamMemberMutation>;
export type AddTeamMemberMutationOptions = Apollo.BaseMutationOptions<AddTeamMemberMutation, AddTeamMemberMutationVariables>;
export const ChangeMemberRoleDocument = gql`
    mutation ChangeMemberRole($input: ChangeMemberRoleInput!) {
  changeMemberRole(input: $input) {
    membership {
      user {
        id
        name
        email
      }
      role
    }
    userErrors {
      field
      message
      code
    }
  }
}
    `;
export type ChangeMemberRoleMutationFn = Apollo.MutationFunction<ChangeMemberRoleMutation, ChangeMemberRoleMutationVariables>;

/**
 * __useChangeMemberRoleMutation__
 *
 * To run a mutation, you first call `useChangeMemberRoleMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useChangeMemberRoleMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [changeMemberRoleMutation, { data, loading, error }] = useChangeMemberRoleMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useChangeMemberRoleMutation(baseOptions?: Apollo.MutationHookOptions<ChangeMemberRoleMutation, ChangeMemberRoleMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<ChangeMemberRoleMutation, ChangeMemberRoleMutationVariables>(ChangeMemberRoleDocument, options);
      }
export type ChangeMemberRoleMutationHookResult = ReturnType<typeof useChangeMemberRoleMutation>;
export type ChangeMemberRoleMutationResult = Apollo.MutationResult<ChangeMemberRoleMutation>;
export type ChangeMemberRoleMutationOptions = Apollo.BaseMutationOptions<ChangeMemberRoleMutation, ChangeMemberRoleMutationVariables>;
export const RemoveTeamMemberDocument = gql`
    mutation RemoveTeamMember($input: RemoveTeamMemberInput!) {
  removeTeamMember(input: $input) {
    success
    userErrors {
      field
      message
      code
    }
  }
}
    `;
export type RemoveTeamMemberMutationFn = Apollo.MutationFunction<RemoveTeamMemberMutation, RemoveTeamMemberMutationVariables>;

/**
 * __useRemoveTeamMemberMutation__
 *
 * To run a mutation, you first call `useRemoveTeamMemberMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useRemoveTeamMemberMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [removeTeamMemberMutation, { data, loading, error }] = useRemoveTeamMemberMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useRemoveTeamMemberMutation(baseOptions?: Apollo.MutationHookOptions<RemoveTeamMemberMutation, RemoveTeamMemberMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<RemoveTeamMemberMutation, RemoveTeamMemberMutationVariables>(RemoveTeamMemberDocument, options);
      }
export type RemoveTeamMemberMutationHookResult = ReturnType<typeof useRemoveTeamMemberMutation>;
export type RemoveTeamMemberMutationResult = Apollo.MutationResult<RemoveTeamMemberMutation>;
export type RemoveTeamMemberMutationOptions = Apollo.BaseMutationOptions<RemoveTeamMemberMutation, RemoveTeamMemberMutationVariables>;
export const MyTeamsDocument = gql`
    query MyTeams {
  myTeams {
    id
    name
    createdAt
  }
}
    `;

/**
 * __useMyTeamsQuery__
 *
 * To run a query within a React component, call `useMyTeamsQuery` and pass it any options that fit your needs.
 * When your component renders, `useMyTeamsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useMyTeamsQuery({
 *   variables: {
 *   },
 * });
 */
export function useMyTeamsQuery(baseOptions?: Apollo.QueryHookOptions<MyTeamsQuery, MyTeamsQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<MyTeamsQuery, MyTeamsQueryVariables>(MyTeamsDocument, options);
      }
export function useMyTeamsLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<MyTeamsQuery, MyTeamsQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<MyTeamsQuery, MyTeamsQueryVariables>(MyTeamsDocument, options);
        }
// @ts-ignore
export function useMyTeamsSuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<MyTeamsQuery, MyTeamsQueryVariables>): Apollo.UseSuspenseQueryResult<MyTeamsQuery, MyTeamsQueryVariables>;
export function useMyTeamsSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<MyTeamsQuery, MyTeamsQueryVariables>): Apollo.UseSuspenseQueryResult<MyTeamsQuery | undefined, MyTeamsQueryVariables>;
export function useMyTeamsSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<MyTeamsQuery, MyTeamsQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<MyTeamsQuery, MyTeamsQueryVariables>(MyTeamsDocument, options);
        }
export type MyTeamsQueryHookResult = ReturnType<typeof useMyTeamsQuery>;
export type MyTeamsLazyQueryHookResult = ReturnType<typeof useMyTeamsLazyQuery>;
export type MyTeamsSuspenseQueryHookResult = ReturnType<typeof useMyTeamsSuspenseQuery>;
export type MyTeamsQueryResult = Apollo.QueryResult<MyTeamsQuery, MyTeamsQueryVariables>;