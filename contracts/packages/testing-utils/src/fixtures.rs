use truerepublic_bindings::*;

pub fn default_domain_response(name: &str) -> DomainResponse {
    DomainResponse {
        name: name.to_string(),
        admin: "cosmos1admin".to_string(),
        member_count: 10,
        treasury: "1000000".to_string(),
        issue_count: 3,
        merkle_root: Some("0xabcdef1234567890".to_string()),
        total_payouts: 5000,
        options: DomainOptionsResponse {
            admin_electable: true,
            anyone_can_join: false,
            only_admin_issues: false,
            coin_burn_required: false,
            voting_mode: 0,
        },
    }
}

pub fn default_domain_members_response(name: &str, count: usize) -> DomainMembersResponse {
    DomainMembersResponse {
        domain_name: name.to_string(),
        members: (0..count).map(|i| format!("cosmos1member{}", i)).collect(),
    }
}

pub fn default_issue_response(name: &str, suggestion_count: usize) -> IssueResponse {
    IssueResponse {
        name: name.to_string(),
        stones: 42,
        suggestion_count: suggestion_count as i64,
        suggestions: (0..suggestion_count)
            .map(|i| SuggestionBrief {
                name: format!("suggestion{}", i),
                creator: format!("cosmos1member{}", i),
                stones: 10 + i as i64,
                color: "green".to_string(),
                score: 5 + i as i64,
            })
            .collect(),
        creation_date: 1700000000,
        external_link: None,
    }
}

pub fn default_purge_schedule_response(name: &str) -> PurgeScheduleResponse {
    PurgeScheduleResponse {
        domain_name: name.to_string(),
        next_purge_time: 1700086400,
        purge_interval: 604800,
        announcement_lead: 86400,
    }
}

pub fn default_treasury_response(name: &str, amount: &str) -> DomainTreasuryResponse {
    DomainTreasuryResponse {
        domain_name: name.to_string(),
        amount: amount.to_string(),
    }
}
