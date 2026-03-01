use cosmwasm_std::{QuerierWrapper, QueryRequest, StdResult};

use crate::query::{
    DomainMembersResponse, DomainResponse, DomainTreasuryResponse, IssueResponse,
    NullifierResponse, PurgeScheduleResponse, SuggestionResponse, TrueRepublicQuery,
};

pub fn query_domain(
    querier: &QuerierWrapper<TrueRepublicQuery>,
    name: &str,
) -> StdResult<DomainResponse> {
    querier.query(&QueryRequest::Custom(TrueRepublicQuery::Domain {
        name: name.to_string(),
    }))
}

pub fn query_domain_members(
    querier: &QuerierWrapper<TrueRepublicQuery>,
    domain_name: &str,
) -> StdResult<DomainMembersResponse> {
    querier.query(&QueryRequest::Custom(TrueRepublicQuery::DomainMembers {
        domain_name: domain_name.to_string(),
    }))
}

pub fn query_issue(
    querier: &QuerierWrapper<TrueRepublicQuery>,
    domain_name: &str,
    issue_name: &str,
) -> StdResult<IssueResponse> {
    querier.query(&QueryRequest::Custom(TrueRepublicQuery::Issue {
        domain_name: domain_name.to_string(),
        issue_name: issue_name.to_string(),
    }))
}

pub fn query_suggestion(
    querier: &QuerierWrapper<TrueRepublicQuery>,
    domain_name: &str,
    issue_name: &str,
    suggestion_name: &str,
) -> StdResult<SuggestionResponse> {
    querier.query(&QueryRequest::Custom(TrueRepublicQuery::Suggestion {
        domain_name: domain_name.to_string(),
        issue_name: issue_name.to_string(),
        suggestion_name: suggestion_name.to_string(),
    }))
}

pub fn query_purge_schedule(
    querier: &QuerierWrapper<TrueRepublicQuery>,
    domain_name: &str,
) -> StdResult<PurgeScheduleResponse> {
    querier.query(&QueryRequest::Custom(TrueRepublicQuery::PurgeSchedule {
        domain_name: domain_name.to_string(),
    }))
}

pub fn query_nullifier(
    querier: &QuerierWrapper<TrueRepublicQuery>,
    domain_name: &str,
    nullifier_hex: &str,
) -> StdResult<NullifierResponse> {
    querier.query(&QueryRequest::Custom(TrueRepublicQuery::Nullifier {
        domain_name: domain_name.to_string(),
        nullifier_hex: nullifier_hex.to_string(),
    }))
}

pub fn query_domain_treasury(
    querier: &QuerierWrapper<TrueRepublicQuery>,
    domain_name: &str,
) -> StdResult<DomainTreasuryResponse> {
    querier.query(&QueryRequest::Custom(TrueRepublicQuery::DomainTreasury {
        domain_name: domain_name.to_string(),
    }))
}
