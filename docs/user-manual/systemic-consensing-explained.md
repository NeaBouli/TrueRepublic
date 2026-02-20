# Systemic Consensing Explained

TrueRepublic uses **Systemic Consensing** instead of traditional Yes/No voting. This document explains the system and how to use it effectively.

## What is Systemic Consensing?

Traditional voting asks: "Do you support this?" (Yes or No).

Systemic Consensing asks: "How much resistance do you have to this?" on a scale from **-5 to +5**.

This produces better outcomes because it:
- Measures the **intensity** of opinions, not just direction
- Finds solutions with the **least overall resistance**
- Makes **minority concerns visible** in the average
- Encourages **compromise** instead of winner-take-all outcomes

## The Rating Scale

```
-5 ────────── 0 ────────── +5
Strong      Neutral      Strong
Resistance               Support
```

| Rating | Meaning | When to Use |
|--------|---------|-------------|
| **-5** | Strongly oppose | "This would cause serious harm" |
| **-4** | Major concerns | "Significant problems need addressing" |
| **-3** | Moderate concerns | "Decent idea but implementation worries me" |
| **-2** | Minor concerns | "Small tweaks needed" |
| **-1** | Slight concerns | "Not perfect but acceptable" |
| **0** | Neutral | "No strong feelings either way" |
| **+1** | Slight preference | "Would be a small improvement" |
| **+2** | Moderate preference | "This would genuinely help" |
| **+3** | Strong preference | "Solves real problems" |
| **+4** | Very strong preference | "Crucial for success" |
| **+5** | Strongest support | "Absolutely necessary, game-changer" |

## Why Not Yes/No?

### Problem with Simple Majority

**Scenario:** 100 members vote on building a new park.
- 51 vote Yes, 49 vote No
- Park is built
- 49% of people are unhappy

### Systemic Consensing Approach

**Same scenario with multiple options:**

| Option | Average Rating | Interpretation |
|--------|---------------|----------------|
| Build new park | +1.2 | Slight support, significant resistance |
| Renovate existing park | +3.5 | Broad strong support |
| Do nothing | -2.1 | Broad resistance |

**Result:** Renovate existing park wins -- it has the **least resistance** and **broadest support**, even though it wasn't anyone's "first choice" in a Yes/No vote.

## How Ratings are Calculated

For each suggestion, TrueRepublic computes:
- **Average Rating** = sum of all ratings / number of raters
- **Approval** = whether average exceeds the domain's approval threshold (default 5%)

### Example

An issue has two suggestions with 5 raters each:

**Suggestion A: "Increase budget by 20%"**
| Rater | Rating | Reason |
|-------|--------|--------|
| Alice | +4 | "Needed for growth" |
| Bob | +3 | "Good investment" |
| Charlie | -2 | "Too aggressive" |
| Diana | +1 | "Okay but cautious" |
| Eve | -4 | "Fiscally irresponsible" |
| **Average** | **+0.4** | Mild support, significant division |

**Suggestion B: "Increase budget by 10%"**
| Rater | Rating | Reason |
|-------|--------|--------|
| Alice | +2 | "Better than nothing" |
| Bob | +3 | "Reasonable compromise" |
| Charlie | +2 | "Sustainable increase" |
| Diana | +3 | "Well-balanced" |
| Eve | +1 | "Acceptable" |
| **Average** | **+2.2** | Strong broad support |

**Winner:** Suggestion B, because it has the least resistance and broadest support.

## Anonymous Ratings (WP S4)

TrueRepublic supports anonymous voting through domain key pairs:

1. Each member generates a **domain-specific key pair**
2. The public key is registered via `join-permission-register`
3. Ratings are submitted with the domain public key, not the member's main address
4. Ratings are **unlinkable** to the member's identity
5. The admin can **purge** the permission register to reset anonymity periodically

This protects voters from social pressure or retaliation.

## How to Rate Effectively

### Step 1: Understand the Proposal
- Read the full issue and suggestion
- Check any external links provided
- Consider who is affected and how

### Step 2: Assess Your Resistance
- What are your specific concerns?
- Are there dealbreakers?
- How strongly do you feel?

### Step 3: Rate Honestly
- Use the **full scale** -- don't cluster around 0
- Base ratings on **merit**, not on who proposed it
- Consider **tradeoffs**, not just whether you personally benefit

### Step 4: Update if Needed
- Ratings can be updated as proposals evolve
- If concerns are addressed, adjust your rating
- If new information emerges, reconsider

## Common Mistakes

| Mistake | Problem | Better Approach |
|---------|---------|-----------------|
| Always rating +5 or -5 | Destroys nuance | Use the full scale honestly |
| Rating based on who proposed | Bias, not merit | Evaluate the idea itself |
| Rating without reading | Uninformed | Read fully or abstain |
| Strategic voting | Undermines consensus | Rate your true resistance |
| Never updating | Ignores new information | Revisit after discussion |

## Comparison with Other Systems

| Feature | Yes/No | Ranked Choice | Systemic Consensing |
|---------|--------|---------------|---------------------|
| Nuance | None | Some | Full (-5 to +5) |
| Minority voice | Ignored | Somewhat | Visible in average |
| Compromise | Winner-take-all | Better | Best (least resistance) |
| Simplicity | Simple | Complex | Intuitive |
| Polarization | Encourages | Reduces | Strongly discourages |
| Multiple options | Separate votes | Single ballot | Rate each independently |

## Tips for Better Outcomes

1. **Start neutral** -- Begin at 0 and adjust based on evidence
2. **Explain extreme ratings** -- If you rate -5 or +5, say why
3. **Consider impact** -- Weight your rating by how much the proposal affects you
4. **Separate person from idea** -- Good ideas can come from anyone
5. **Stay engaged** -- Update your ratings as discussion progresses

## Further Reading

- [Governance Tutorial](governance-tutorial.md) -- Apply this knowledge
- [Stones Voting Guide](stones-voting-guide.md) -- Complementary voting mechanism
- TrueRepublic Whitepaper S3: Governance Mechanisms
