use cosmwasm_std::testing::{MockApi, MockQuerier, MockQuerierCustomHandlerResult, MockStorage};
use cosmwasm_std::{
    to_json_binary, ContractResult, OwnedDeps, Querier, QuerierResult, SystemResult,
};
use truerepublic_bindings::*;

pub struct TrueRepublicMockQuerier {
    pub base: MockQuerier<TrueRepublicQuery>,
}

impl Querier for TrueRepublicMockQuerier {
    fn raw_query(&self, bin_request: &[u8]) -> QuerierResult {
        self.base.raw_query(bin_request)
    }
}

impl TrueRepublicMockQuerier {
    pub fn new(balances: &[(&str, &[cosmwasm_std::Coin])]) -> Self {
        let querier = MockQuerier::<TrueRepublicQuery>::new(balances).with_custom_handler(
            |query| -> MockQuerierCustomHandlerResult {
                match query {
                    TrueRepublicQuery::Domain { name } => SystemResult::Ok(ContractResult::Ok(
                        to_json_binary(&DomainResponse {
                            name: name.clone(),
                            admin: "admin1".to_string(),
                            member_count: 10,
                            treasury: "1000000".to_string(),
                            issue_count: 3,
                            merkle_root: Some("0xabcdef".to_string()),
                            total_payouts: 5000,
                            options: DomainOptionsResponse {
                                admin_electable: true,
                                anyone_can_join: false,
                                only_admin_issues: false,
                                coin_burn_required: false,
                                voting_mode: 0,
                            },
                        })
                        .unwrap(),
                    )),
                    TrueRepublicQuery::DomainMembers { domain_name } => {
                        SystemResult::Ok(ContractResult::Ok(
                            to_json_binary(&DomainMembersResponse {
                                domain_name: domain_name.clone(),
                                members: vec!["member1".to_string(), "member2".to_string()],
                            })
                            .unwrap(),
                        ))
                    }
                    TrueRepublicQuery::Issue {
                        domain_name: _,
                        issue_name,
                    } => SystemResult::Ok(ContractResult::Ok(
                        to_json_binary(&IssueResponse {
                            name: issue_name.clone(),
                            stones: 42,
                            suggestion_count: 2,
                            suggestions: vec![SuggestionBrief {
                                name: "suggestion1".to_string(),
                                creator: "member1".to_string(),
                                stones: 20,
                                color: "green".to_string(),
                                score: 15,
                            }],
                            creation_date: 1700000000,
                            external_link: None,
                        })
                        .unwrap(),
                    )),
                    TrueRepublicQuery::Suggestion { .. } => SystemResult::Ok(ContractResult::Ok(
                        to_json_binary(&SuggestionResponse {
                            name: "suggestion1".to_string(),
                            creator: "member1".to_string(),
                            stones: 20,
                            color: "green".to_string(),
                            rating_count: 5,
                            score: 15,
                            dwell_time: 86400,
                            creation_date: 1700000000,
                            external_link: None,
                            delete_votes: 0,
                        })
                        .unwrap(),
                    )),
                    TrueRepublicQuery::PurgeSchedule { domain_name } => {
                        SystemResult::Ok(ContractResult::Ok(
                            to_json_binary(&PurgeScheduleResponse {
                                domain_name: domain_name.clone(),
                                next_purge_time: 1700086400,
                                purge_interval: 604800,
                                announcement_lead: 86400,
                            })
                            .unwrap(),
                        ))
                    }
                    TrueRepublicQuery::Nullifier {
                        domain_name: _,
                        nullifier_hex: _,
                    } => SystemResult::Ok(ContractResult::Ok(
                        to_json_binary(&NullifierResponse { used: false }).unwrap(),
                    )),
                    TrueRepublicQuery::DomainTreasury { domain_name } => {
                        SystemResult::Ok(ContractResult::Ok(
                            to_json_binary(&DomainTreasuryResponse {
                                domain_name: domain_name.clone(),
                                amount: "1000000pnyx".to_string(),
                            })
                            .unwrap(),
                        ))
                    }
                }
            },
        );
        Self { base: querier }
    }
}

pub fn mock_dependencies_with_truerepublic(
    balances: &[(&str, &[cosmwasm_std::Coin])],
) -> OwnedDeps<MockStorage, MockApi, TrueRepublicMockQuerier, TrueRepublicQuery> {
    OwnedDeps {
        storage: MockStorage::default(),
        api: MockApi::default(),
        querier: TrueRepublicMockQuerier::new(balances),
        custom_query_type: std::marker::PhantomData,
    }
}
