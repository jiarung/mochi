package errors

import (
	"net/http"
)

// Error code constants.
const (
	UnexpectedError     = "unexpected_error"
	AuthenticationError = "authentication_error"
	OriginError         = "origin_error"
	InvalidPayLoad      = "invalid_payload"
	ParseJSONError      = "parse_json_error"
	ParameterError      = "parameter_error"

	InvalidIPAddress            = "invalid_ip_address"
	InvalidCode                 = "invalid_code"
	ReachRetryLimit             = "reach_retry_limit"
	RegistrationAlreadyVerified = "registration_already_verified"
	RegistrationCancelled       = "registration_cancelled"
	PendingDeviceVerification   = "pending_device_verification"
	PendingEmailVerification    = "pending_email_verification"
	AccessTokenAlreadyGenerate  = "access_token_already_generate"
	AccountLocked               = "account_locked"
	AccountDisabled             = "account_disabled"
	AccountDeleted              = "account_deleted"
	UnauthorizedScope           = "unauthorized_scope"
	InvalidPassword             = "invalid_password"
	StatusConflict              = "status_conflict"
	TryAgainLater               = "try_again_later"
	ServiceDown                 = "service_down"
	EmailExist                  = "email_exist"
	PasswordLowStrength         = "pwd_low_strength"
	EmailVerified               = "email_verified"
	EmailNotExist               = "email_not_exist"
	InvalidToken                = "invalid_token"
	TokenExpired                = "token_expired"
	InvalidAccountType          = "invalid_account_type"
	CorporationIsEmpty          = "corporation_is_empty"
	TokenNotExist               = "token_not_exist"
	EmailAlreadyVerified        = "email_already_verified"
	UserNotExist                = "user_not_found"
	InvalidNonce                = "invalid_nonce"
	AlreadyConfirmed            = "already_confirmed"
	AlreadyVerified             = "already_verified"
	NewRequestFound             = "new_request_found"

	WaitForCooldownTime      = "wait_for_cooldown_time_%s"
	BatchInternalTransferErr = "batch_internal_transfer_err_%s"

	InvalidCurrency          = "invalid_currency"
	InvalidBlockchain        = "invalid_blockchain"
	FundingDisabled          = "funding_disabled"
	MarginDisabled           = "margin_disabled"
	InvalidTradeID           = "invalid_trade_id"
	TradeNotFound            = "trade_not_found"
	TradingPairNotFound      = "trading_pair_not_found"
	TradingPairBlackListed   = "trading_pair_blacklisted"
	BalanceLocked            = "balance_locked"
	InvalidOrderSize         = "invalid_order_size"
	InvalidOrderPrice        = "invalid_order_price"
	OrderNotFound            = "order_not_found"
	InsufficientBalance      = "insufficient_balance"
	BalancesNotFound         = "balances_not_found"
	WalletsNotFound          = "wallets_not_found"
	WalletsExist             = "wallets_exist"
	LedgersNotFound          = "ledgers_not_found"
	ExchangeRateNotFound     = "exchange_rate_not_found"
	IDTypeMapNotFound        = "id_type_map_not_found"
	DepositNotFound          = "deposit_not_found"
	WithdrawalFrozen         = "withdrawal_frozen"
	WithdrawalNotFound       = "withdrawal_not_found"
	WithdrawalLimitsNotFound = "withdrawal_limits_not_found"
	InvalidAddress           = "invalid_address"
	InvalidWithdrawalStatus  = "invalid_withdrawal_status"
	InvalidWithdrawalID      = "invalid_withdrawal_id"
	InvalidDepositStatus     = "invalid_deposit_status"
	InvalidDepositID         = "invalid_deposit_id"
	NotBelongToUser          = "not_belong_to_user"
	TickerNotFound           = "ticker_not_found"
	CandleNotFound           = "candle_not_found"
	DBError                  = "database_or_query_error"
	DeviceNotFound           = "device_not_found"
	FiatDepositNotFound      = "fiat_deposit_not_found"
	FiatWithdrawalNotFound   = "fiat_withdrawal_not_found"
	CommitteeMotionNotFound  = "committee_motion_not_found"

	SlotMachineTokenNotFound  = "slot_machine_token_not_found"
	SlotMachineRewardNotFound = "slot_machine_reward_not_found"
	SlotMachineNotAvailable   = "slot_machine_not_available"

	SlotMachineNoTrading             = "slot_machine_no_trading"
	SlotMachineAlreadyReceivedReward = "slot_machine_already_received_reward"

	ExceedMaxTokenNumber = "exceed_max_token_number"

	InvalidGDPRStatus       = "invalid_gdpr_status"
	DuplicateGDPRSubmission = "duplicate_gdpr_submission"

	TierLevelNotFound      = "tier_level_not_found"
	InvalidKYCForm         = "invalid_kyc_form"
	DuplicateKYCSubmission = "duplicate_kyc_submission"
	KYCMotionNotFount      = "kyc_motion_not_found"
	DuplicateKYCMotion     = "duplicate_kyc_motion"

	InvalidKYCTierAction          = "invalid_kyc_tier_action"
	InvalidKYCTierStatus          = "invalid_kyc_tier_status"
	InvalidAMLStatus              = "invalid_aml_status"
	InvalidAMLInvestigationAction = "invalid_aml_investigating_action"

	InvalidPathParameter  = "invalid_path_parameter"
	InvalidQueryParameter = "invalid_query_parameter"
	NotPrivilegedIP       = "not_privileged_ip"
	ResourceNotFound      = "resource_not_found"
	ForbiddenAnnounceTime = "forbidden_announce_time"
	ForbiddenEndTime      = "forbidden_end_time"
	ForbiddenModification = "forbidden_modification"
	ForbiddenFeature      = "forbidden_feature"

	SkipOrReverseKYCTier     = "skip_reverse_kyc_tier"
	InvalidKYCStatus         = "invalid_kyc_status"
	ActionOnNotQueuedKYCTier = "action_on_not_queued_kyc_tier"
	InvalidAMLAction         = "invalid_aml_action"
	InvalidKYCAction         = "invalid_kyc_action"

	NotKYCAuditor  = "not_kyc_auditor"
	NotDeveloper   = "not_developer"
	NoExpectedRole = "no_expected_role"

	SubjectTooLong  = "subject_too_long"
	ContentTooLong  = "content_too_long"
	CategoryTooLong = "category_too_long"
	InvalidPriority = "invalid_priority"
	InvalidLocale   = "invalid_locale"

	InvalidContentType = "invalid_content_type"
	InvalidFilename    = "invalid_filename"
	FileTooLarge       = "file_too_large"
	CacheKeyNotExists  = "cache_key_not_exists"

	DailyWithdrawLimitExceeded   = "daily_withdraw_limit_exceeded"
	MonthlyWithdrawLimitExceeded = "monthly_withdraw_limit_exceeded"
	AmountLessThanFee            = "amount_less_than_fee"

	NotInCampaignPeriod               = "not_in_campaign_period"
	NotInRedemptionPeriod             = "not_in_redemption_period"
	DuplicateCampaignVote             = "duplicate_campaign_vote"
	InsufficientCOB                   = "insufficient_cob"
	NoCMTToConsume                    = "no_cmt_to_consume"
	DuplicateRedeemedTokenConsumption = "duplicate_redeemed_token_consumption"

	NotLotteryWinner            = "not_lottery_winner"
	DuplicateLotteryConsumption = "duplicate_lottery_consumption"

	NotCampaignWinner                  = "not_campaign_winner"
	DuplicateCampaignRewardConsumption = "duplicate_campaign_reward_consumption"

	FiatDepositDuplicate = "fiat_deposit_duplicate"
	FiatDepositFailed    = "fiat_deposit_failed"
	FiatRefundFailed     = "fiat_refund_failed"

	FiatDepositExceedLimit    = "fiat_deposit_exceed_limit"
	FiatWithdrawalExceedLimit = "fiat_withdrawal_exceed_limit"
	ExceedLimit               = "exceed_daily_limit"

	ExistingTransaction = "existing_transaction"

	StatusMethodNotAllowed = "status_method_not_allow"

	InvalidReferrerID = "invalid_referrer_id"

	OAuth2EmptyState     = "oauth2_empty_state"
	OAuth2DuplicateState = "oauth2_duplicate_state"

	OAuth2InvalidRequest          = "oauth2_invalid_request"
	OAuth2UnauthorizedClient      = "oauth2_unauthorized_client"
	OAuth2AccessDenied            = "oauth2_access_denied"
	OAuth2UnsupportedResponseType = "oauth2_unsupported_response_type"
	OAuth2ServerError             = "oauth2_server_error"
	OAuth2TemporarilyUnavailable  = "oauth2_temporarily_unavailable"

	InvalidPreferenceKey = "invalid_preference_key"

	APIUnderDevelopment = "api_under_development"

	PriceAlertNotFound    = "price_alert_not_found"
	PriceAlertStatusError = "price_alert_status_error"
	PriceAlertTooMany     = "price_alert_too_many"

	PromoCodeUsed         = "promo_code_used"
	PromoCodeExpired      = "promo_code_expired"
	PromoCodeUserRedeemed = "promo_code_user_redeemed"
	PromoCodeSameEvent    = "promo_code_same_event"
	PromoCodeNotExist     = "promo_code_not_exist"

	ManualReviewNoBudget = "manual_review_no_budget"
	ManualReviewResubmit = "manual_review_resubmit"

	InvalidWithdrawalMemo = "invalid_withdrawal_memo"
	InvalidWithdrawalTag  = "invalid_withdrawal_tag"

	ServiceNotAlive            = "service_not_alive"
	WaitForPaymentOrderExceeds = "wait_for_payment_order_exceeds"
	UnderMinLimit              = "under_min_limit_per_order"
	OverMaxLimit               = "over_max_limit_per_order"

	NotWithinSaleTime            = "not_within_sale_time"
	CurrencyNotEnough            = "currency_not_enough"
	PointNotEnough               = "point_not_enough"
	KYCLevelNotEnough            = "kyc_level_not_enough"
	UserLimitExceed              = "user_limit_exceeded"
	ItemSoldOut                  = "item_sold_out"
	BuySmallerTradeContestFactor = "buy_smaller_trade_contest_factor"

	ApprovedRequestExisted = "approved_request_existed"

	KYCProjectNotEnabled = "kyc_project_not_enabled"
	KYCProjectNotJoined  = "kyc_project_not_joined"

	AccountNotFound         = "account_not_found"
	BlockNotFound           = "block_not_found"
	TransactionNotFound     = "transaction_not_found"
	ContractNotFound        = "contract_not_found"
	ContractAlreadyVerified = "contract_already_verified"
	ContractCompileError    = "contract_compile_error"
	ContractVerifyError     = "contract_verify_error"
	TokenNotFound           = "token_not_found"

	GetEOSAccountError        = "get_eos_account_error"
	CreateEOSAccountError     = "create_eos_account_error"
	GetEOSRequiredAmountError = "get_eos_required_amount_error"

	FundsRaisingDepositLessThanMinAmount = "funds_raising_deposit_less_than_min_amount"
	FundsRaisingDepositMoreThanCapacity  = "funds_raising_deposit_more_than_capacity"
	FundsRaisingDepositInvalidPeriod     = "funds_raising_deposit_invalid_period"

	CoinOfferingDepositMoreThanKYCLimit = "coin_offering_deposit_more_than_kyc_limit"
	CoinOfferingDepositMoreThanLimit    = "coin_offering_deposit_more_than_limit"
	CoinOfferingDepositRateNotMatch     = "coin_offering_deposit_rate_not_match"
	CoinOfferingBlackListed             = "coin_offering_blacklisted"

	AuditCommitteeProposeFailed = "audit_committee_propose_failed"
	AuditCommitteeVoteFailed    = "audit_committee_vote_failed"

	URLParseError     = "url_parse_error"
	TweetContentError = "tweet_content_error"
	TweetUsed         = "tweet_used"
)

var errorCodeMap = map[string]int{
	UnexpectedError:             http.StatusInternalServerError,
	AuthenticationError:         http.StatusUnauthorized,
	OriginError:                 http.StatusUnauthorized,
	ParseJSONError:              http.StatusBadRequest,
	ParameterError:              http.StatusBadRequest,
	TryAgainLater:               http.StatusTooManyRequests,
	ServiceDown:                 http.StatusTooManyRequests,
	EmailExist:                  http.StatusBadRequest,
	PasswordLowStrength:         http.StatusBadRequest,
	EmailVerified:               http.StatusBadRequest,
	EmailNotExist:               http.StatusBadRequest,
	InvalidToken:                http.StatusBadRequest,
	TokenExpired:                http.StatusBadRequest,
	InvalidAccountType:          http.StatusBadRequest,
	InvalidPayLoad:              http.StatusBadRequest,
	CorporationIsEmpty:          http.StatusBadRequest,
	TokenNotExist:               http.StatusBadRequest,
	EmailAlreadyVerified:        http.StatusBadRequest,
	UserNotExist:                http.StatusBadRequest,
	InvalidNonce:                http.StatusConflict,
	InvalidCurrency:             http.StatusBadRequest,
	InvalidBlockchain:           http.StatusBadRequest,
	FundingDisabled:             http.StatusBadRequest,
	MarginDisabled:              http.StatusBadRequest,
	InvalidTradeID:              http.StatusBadRequest,
	TradeNotFound:               http.StatusNotFound,
	TradingPairNotFound:         http.StatusNotFound,
	TradingPairBlackListed:      http.StatusForbidden,
	BalanceLocked:               http.StatusBadRequest,
	InvalidOrderSize:            http.StatusBadRequest,
	InvalidOrderPrice:           http.StatusBadRequest,
	InvalidAMLAction:            http.StatusBadRequest,
	InvalidKYCAction:            http.StatusBadRequest,
	OrderNotFound:               http.StatusNotFound,
	NotBelongToUser:             http.StatusForbidden,
	StatusConflict:              http.StatusConflict,
	AccountLocked:               http.StatusForbidden,
	AccountDisabled:             http.StatusForbidden,
	AccountDeleted:              http.StatusForbidden,
	AccessTokenAlreadyGenerate:  http.StatusBadRequest,
	PendingEmailVerification:    http.StatusBadRequest,
	RegistrationCancelled:       http.StatusBadRequest,
	RegistrationAlreadyVerified: http.StatusBadRequest,
	ReachRetryLimit:             http.StatusBadRequest,
	UnauthorizedScope:           http.StatusForbidden,
	InvalidPassword:             http.StatusBadRequest,
	InvalidCode:                 http.StatusBadRequest,
	InvalidIPAddress:            http.StatusBadRequest,
	PendingDeviceVerification:   http.StatusBadRequest,
	InsufficientBalance:         http.StatusBadRequest,
	DepositNotFound:             http.StatusNotFound,
	WithdrawalFrozen:            http.StatusForbidden,
	WithdrawalNotFound:          http.StatusNotFound,
	WithdrawalLimitsNotFound:    http.StatusNotFound,
	InvalidAddress:              http.StatusBadRequest,
	InvalidWithdrawalStatus:     http.StatusBadRequest,
	InvalidWithdrawalID:         http.StatusBadRequest,
	InvalidDepositStatus:        http.StatusBadRequest,
	InvalidDepositID:            http.StatusBadRequest,
	BalancesNotFound:            http.StatusNotFound,
	WalletsNotFound:             http.StatusNotFound,
	WalletsExist:                http.StatusConflict,
	TickerNotFound:              http.StatusNotFound,
	CandleNotFound:              http.StatusNotFound,
	IDTypeMapNotFound:           http.StatusNotFound,
	ExchangeRateNotFound:        http.StatusBadRequest,
	DBError:                     http.StatusBadRequest,
	DeviceNotFound:              http.StatusNotFound,
	FiatDepositNotFound:         http.StatusNotFound,
	FiatWithdrawalNotFound:      http.StatusNotFound,
	CommitteeMotionNotFound:     http.StatusNotFound,
	AlreadyConfirmed:            http.StatusBadRequest,
	AlreadyVerified:             http.StatusBadRequest,
	NewRequestFound:             http.StatusBadRequest,

	WaitForCooldownTime:      http.StatusTooManyRequests,
	BatchInternalTransferErr: http.StatusBadRequest,

	SlotMachineTokenNotFound:  http.StatusNotFound,
	SlotMachineRewardNotFound: http.StatusNotFound,
	SlotMachineNotAvailable:   http.StatusBadRequest,

	SlotMachineNoTrading:             http.StatusBadRequest,
	SlotMachineAlreadyReceivedReward: http.StatusBadRequest,

	ExceedMaxTokenNumber: http.StatusBadRequest,

	InvalidGDPRStatus:       http.StatusBadRequest,
	DuplicateGDPRSubmission: http.StatusBadRequest,

	TierLevelNotFound:      http.StatusNotFound,
	InvalidKYCForm:         http.StatusBadRequest,
	DuplicateKYCSubmission: http.StatusBadRequest,
	KYCMotionNotFount:      http.StatusBadRequest,
	DuplicateKYCMotion:     http.StatusBadRequest,
	InvalidKYCStatus:       http.StatusBadRequest,

	InvalidKYCTierAction:          http.StatusBadRequest,
	InvalidAMLStatus:              http.StatusBadRequest,
	InvalidAMLInvestigationAction: http.StatusBadRequest,

	InvalidPathParameter:  http.StatusBadRequest,
	InvalidQueryParameter: http.StatusBadRequest,
	NotPrivilegedIP:       http.StatusForbidden,
	ResourceNotFound:      http.StatusNotFound,
	ForbiddenAnnounceTime: http.StatusForbidden,
	ForbiddenEndTime:      http.StatusForbidden,
	ForbiddenModification: http.StatusForbidden,
	ForbiddenFeature:      http.StatusForbidden,

	SkipOrReverseKYCTier:     http.StatusBadRequest,
	ActionOnNotQueuedKYCTier: http.StatusBadRequest,

	NotKYCAuditor: http.StatusBadRequest,
	NotDeveloper:  http.StatusBadRequest,

	SubjectTooLong:  http.StatusBadRequest,
	ContentTooLong:  http.StatusBadRequest,
	CategoryTooLong: http.StatusBadRequest,
	InvalidPriority: http.StatusBadRequest,
	InvalidLocale:   http.StatusBadRequest,

	InvalidFilename:    http.StatusBadRequest,
	InvalidContentType: http.StatusBadRequest,
	FileTooLarge:       http.StatusBadRequest,
	CacheKeyNotExists:  http.StatusBadRequest,

	DailyWithdrawLimitExceeded:   http.StatusBadRequest,
	MonthlyWithdrawLimitExceeded: http.StatusBadRequest,
	AmountLessThanFee:            http.StatusBadRequest,

	NotInCampaignPeriod:               http.StatusBadRequest,
	NotInRedemptionPeriod:             http.StatusBadRequest,
	DuplicateCampaignVote:             http.StatusBadRequest,
	InsufficientCOB:                   http.StatusBadRequest,
	NoCMTToConsume:                    http.StatusBadRequest,
	DuplicateRedeemedTokenConsumption: http.StatusBadRequest,

	NotLotteryWinner:                   http.StatusBadRequest,
	DuplicateLotteryConsumption:        http.StatusBadRequest,
	NotCampaignWinner:                  http.StatusBadRequest,
	DuplicateCampaignRewardConsumption: http.StatusBadRequest,

	FiatDepositDuplicate: 250,
	FiatDepositFailed:    543,

	FiatDepositExceedLimit:    http.StatusBadRequest,
	FiatWithdrawalExceedLimit: http.StatusBadRequest,
	ExceedLimit:               http.StatusBadRequest,

	ExistingTransaction: http.StatusBadRequest,

	StatusMethodNotAllowed: http.StatusBadRequest,

	InvalidReferrerID: http.StatusBadRequest,

	OAuth2EmptyState:     http.StatusForbidden,
	OAuth2DuplicateState: http.StatusForbidden,

	OAuth2InvalidRequest:          http.StatusBadRequest,
	OAuth2UnauthorizedClient:      http.StatusUnauthorized,
	OAuth2AccessDenied:            http.StatusForbidden,
	OAuth2UnsupportedResponseType: http.StatusNotImplemented,
	OAuth2ServerError:             http.StatusInternalServerError,
	OAuth2TemporarilyUnavailable:  http.StatusServiceUnavailable,

	InvalidPreferenceKey: http.StatusNotFound,

	APIUnderDevelopment: http.StatusNotFound,

	PriceAlertNotFound:    http.StatusNotFound,
	PriceAlertStatusError: http.StatusBadRequest,
	PriceAlertTooMany:     http.StatusBadRequest,

	PromoCodeUsed:         http.StatusBadRequest,
	PromoCodeExpired:      http.StatusBadRequest,
	PromoCodeUserRedeemed: http.StatusBadRequest,
	PromoCodeSameEvent:    http.StatusBadRequest,
	PromoCodeNotExist:     http.StatusBadRequest,

	ManualReviewNoBudget: http.StatusBadRequest,
	ManualReviewResubmit: http.StatusBadRequest,

	InvalidWithdrawalMemo: http.StatusBadRequest,
	InvalidWithdrawalTag:  http.StatusBadRequest,

	ServiceNotAlive:            http.StatusServiceUnavailable,
	WaitForPaymentOrderExceeds: http.StatusBadRequest,
	UnderMinLimit:              http.StatusBadRequest,
	OverMaxLimit:               http.StatusBadRequest,

	NotWithinSaleTime:            http.StatusBadRequest,
	CurrencyNotEnough:            http.StatusBadRequest,
	PointNotEnough:               http.StatusBadRequest,
	KYCLevelNotEnough:            http.StatusBadRequest,
	UserLimitExceed:              http.StatusBadRequest,
	ItemSoldOut:                  http.StatusBadRequest,
	BuySmallerTradeContestFactor: http.StatusBadRequest,

	ApprovedRequestExisted: http.StatusBadRequest,

	KYCProjectNotEnabled: http.StatusBadRequest,
	KYCProjectNotJoined:  http.StatusBadRequest,

	AccountNotFound:         http.StatusNotFound,
	BlockNotFound:           http.StatusNotFound,
	TransactionNotFound:     http.StatusNotFound,
	ContractNotFound:        http.StatusNotFound,
	ContractAlreadyVerified: http.StatusBadRequest,
	ContractCompileError:    http.StatusBadRequest,
	ContractVerifyError:     http.StatusBadRequest,
	TokenNotFound:           http.StatusNotFound,

	GetEOSAccountError:        http.StatusInternalServerError,
	CreateEOSAccountError:     http.StatusInternalServerError,
	GetEOSRequiredAmountError: http.StatusInternalServerError,

	FundsRaisingDepositLessThanMinAmount: http.StatusBadRequest,
	FundsRaisingDepositMoreThanCapacity:  http.StatusBadRequest,
	FundsRaisingDepositInvalidPeriod:     http.StatusBadRequest,

	CoinOfferingDepositMoreThanKYCLimit: http.StatusBadRequest,
	CoinOfferingDepositMoreThanLimit:    http.StatusBadRequest,
	CoinOfferingDepositRateNotMatch:     http.StatusBadRequest,
	CoinOfferingBlackListed:             http.StatusForbidden,

	AuditCommitteeProposeFailed: http.StatusBadRequest,
	AuditCommitteeVoteFailed:    http.StatusBadRequest,

	URLParseError:     http.StatusBadRequest,
	TweetContentError: http.StatusBadRequest,
	TweetUsed:         http.StatusBadRequest,
}

// HTTPStatus returns HTTP status from errorCodeMap.
func HTTPStatus(code string) int {
	if status, ok := errorCodeMap[code]; ok {
		return status
	}
	return errorCodeMap[UnexpectedError]
}
