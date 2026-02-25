package truedemocracy

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CheckAndExecuteBigPurges runs in EndBlock to check all domains for
// scheduled Big Purge execution (WP S4). Lightweight: only compares
// int64 timestamps, no heavy computation.
func (k Keeper) CheckAndExecuteBigPurges(ctx sdk.Context) {
	now := ctx.BlockTime().Unix()

	// Collect domain names first to avoid modifying store during iteration.
	var domainNames []string
	k.IterateDomains(ctx, func(d Domain) bool {
		domainNames = append(domainNames, d.Name)
		return false
	})

	for _, name := range domainNames {
		k.checkDomainPurge(ctx, name, now)
	}
}

// checkDomainPurge checks a single domain's purge schedule and acts accordingly.
func (k Keeper) checkDomainPurge(ctx sdk.Context, domainName string, now int64) {
	schedule, exists := k.GetBigPurgeSchedule(ctx, domainName)
	if !exists {
		return
	}

	// Check if purge time has been reached.
	if now >= schedule.NextPurgeTime {
		k.executeBigPurge(ctx, domainName)

		// Reschedule next purge from current time.
		schedule.NextPurgeTime = now + schedule.PurgeInterval
		k.SetBigPurgeSchedule(ctx, schedule)

		// Clear announcement flag for this cycle.
		k.clearPurgeAnnounced(ctx, domainName)
		return
	}

	// Check if we should emit an announcement (within AnnouncementLead of purge).
	announcementTime := schedule.NextPurgeTime - schedule.AnnouncementLead
	if now >= announcementTime && !k.hasPurgeBeenAnnounced(ctx, domainName) {
		ctx.EventManager().EmitEvent(sdk.NewEvent(
			"big_purge_announcement",
			sdk.NewAttribute("domain", domainName),
			sdk.NewAttribute("purge_time", fmt.Sprintf("%d", schedule.NextPurgeTime)),
			sdk.NewAttribute("block_height", fmt.Sprintf("%d", ctx.BlockHeight())),
		))
		k.setPurgeAnnounced(ctx, domainName)
	}
}

// executeBigPurge clears a domain's permission register. This is system-initiated
// from EndBlock so no admin auth is required. Member list stays intact (WP S4).
func (k Keeper) executeBigPurge(ctx sdk.Context, domainName string) {
	domain, found := k.GetDomain(ctx, domainName)
	if !found {
		return
	}

	domain.PermissionReg = []string{}

	store := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&domain)
	store.Set([]byte("domain:"+domainName), bz)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"big_purge_executed",
		sdk.NewAttribute("domain", domainName),
		sdk.NewAttribute("block_height", fmt.Sprintf("%d", ctx.BlockHeight())),
	))
}

// ---------- Announcement tracking ----------

func purgeAnnouncedKey(domainName string) []byte {
	return []byte("purge-announced:" + domainName)
}

func (k Keeper) hasPurgeBeenAnnounced(ctx sdk.Context, domainName string) bool {
	store := ctx.KVStore(k.StoreKey)
	return store.Has(purgeAnnouncedKey(domainName))
}

func (k Keeper) setPurgeAnnounced(ctx sdk.Context, domainName string) {
	store := ctx.KVStore(k.StoreKey)
	store.Set(purgeAnnouncedKey(domainName), []byte{1})
}

func (k Keeper) clearPurgeAnnounced(ctx sdk.Context, domainName string) {
	store := ctx.KVStore(k.StoreKey)
	store.Delete(purgeAnnouncedKey(domainName))
}
