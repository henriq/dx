// Shared logic for the Cleanup prereleases workflow.
// Invoked from both the `smoke` job (dryRun=true, preview on PRs) and the
// `cleanup` job (always dryRun=false for real execution).

const PRERELEASE_TAG_PATTERN = /^v\d+\.\d+\.\d+-prerelease\+[0-9a-f]{7,}$/;
const MUTATING_CALL_THROTTLE_MS = 250;

const sleep = (milliseconds) => new Promise(resolve => setTimeout(resolve, milliseconds));

const isAlreadyGoneError = (error) => error.status === 404 || error.status === 422;

function comparePrereleasesNewestFirst(left, right) {
  const leftTime = new Date(left.published_at || left.created_at);
  const rightTime = new Date(right.published_at || right.created_at);
  const timeDiff = rightTime - leftTime;
  return timeDiff !== 0 ? timeDiff : right.id - left.id;
}

async function fetchPrereleasesNewestFirst({ github, owner, repo }) {
  const releases = await github.paginate(
    github.rest.repos.listReleases,
    { owner, repo, per_page: 100 }
  );
  return releases
    .filter(release => release.prerelease)
    .sort(comparePrereleasesNewestFirst);
}

async function fetchPrereleaseTagNames({ github, owner, repo }) {
  const tagRefs = await github.paginate(
    github.rest.git.listMatchingRefs,
    { owner, repo, ref: 'tags/v', per_page: 100 }
  );
  return tagRefs
    .map(tagRef => tagRef.ref.replace(/^refs\/tags\//, ''))
    .filter(tagName => PRERELEASE_TAG_PATTERN.test(tagName));
}

async function deleteRelease({ github, owner, repo, releaseId }) {
  await github.rest.repos.deleteRelease({ owner, repo, release_id: releaseId });
}

async function deleteTagRef({ github, owner, repo, tagName }) {
  await github.request('DELETE /repos/{owner}/{repo}/git/refs/tags/{tag}', {
    owner, repo, tag: tagName,
  });
}

function recordDeleteError({ core, error, subject }) {
  if (isAlreadyGoneError(error)) {
    const capitalized = subject.charAt(0).toUpperCase() + subject.slice(1);
    core.warning(`${capitalized} already deleted or not found`);
    return { alreadyGone: true };
  }
  core.error(`Failed to delete ${subject} (status ${error.status}): ${error.message}`);
  return { alreadyGone: false };
}

async function deleteOnePrereleaseWithTag({ github, core, owner, repo, release }) {
  console.log(`Deleting release: ${release.tag_name}`);
  try {
    await deleteRelease({ github, owner, repo, releaseId: release.id });
  } catch (error) {
    const { alreadyGone } = recordDeleteError({ core, error, subject: `release ${release.tag_name}` });
    // The orphan sweep is the recovery path when an already-gone release skips the tag delete.
    return { deleted: 0, failed: alreadyGone ? 0 : 1 };
  }

  let failed = 0;
  try {
    await deleteTagRef({ github, owner, repo, tagName: release.tag_name });
    console.log(`Deleted tag: ${release.tag_name}`);
  } catch (error) {
    const { alreadyGone } = recordDeleteError({ core, error, subject: `tag ${release.tag_name}` });
    if (!alreadyGone) failed = 1;
  }

  return { deleted: 1, failed };
}

async function deleteOldPrereleases({ github, core, owner, repo, prereleases, dryRun }) {
  const toDelete = prereleases.slice(1);
  console.log(`${dryRun ? '[dry-run] Would delete' : 'Deleting'} ${toDelete.length} old prerelease(s)`);

  if (dryRun) {
    for (const release of toDelete) {
      console.log(`[dry-run] Would delete: ${release.tag_name}`);
    }
    return { deleted: toDelete.length, failed: 0 };
  }

  let deleted = 0;
  let failed = 0;
  for (const release of toDelete) {
    try {
      const result = await deleteOnePrereleaseWithTag({ github, core, owner, repo, release });
      deleted += result.deleted;
      failed += result.failed;
    } finally {
      await sleep(MUTATING_CALL_THROTTLE_MS);
    }
  }
  return { deleted, failed };
}

async function sweepOrphanTags({ github, core, owner, repo, dryRun, attachedTagNames }) {
  const allPrereleaseTagNames = await fetchPrereleaseTagNames({ github, owner, repo });
  const orphanTagNames = allPrereleaseTagNames.filter(tagName => !attachedTagNames.has(tagName));

  if (orphanTagNames.length === 0) {
    console.log(dryRun ? '[dry-run] No orphan prerelease tags to delete' : 'No orphan prerelease tags to delete');
    return { deleted: 0, failed: 0 };
  }

  console.log(`${dryRun ? '[dry-run] Would delete' : 'Deleting'} ${orphanTagNames.length} orphan prerelease tag(s)`);

  if (dryRun) {
    for (const tagName of orphanTagNames) {
      console.log(`[dry-run] Would delete orphan tag: ${tagName}`);
    }
    return { deleted: orphanTagNames.length, failed: 0 };
  }

  let deleted = 0;
  let failed = 0;
  for (const tagName of orphanTagNames) {
    try {
      await deleteTagRef({ github, owner, repo, tagName });
      console.log(`Deleted orphan tag: ${tagName}`);
      deleted++;
    } catch (error) {
      const { alreadyGone } = recordDeleteError({ core, error, subject: `orphan tag ${tagName}` });
      if (!alreadyGone) failed++;
    } finally {
      await sleep(MUTATING_CALL_THROTTLE_MS);
    }
  }
  return { deleted, failed };
}

async function writeSummary({ core, dryRun, keptTagName, prereleaseDeleted, orphanDeleted, failed }) {
  await core.summary
    .addHeading(`Prerelease Cleanup${dryRun ? ' (dry-run)' : ''}`)
    .addList([
      `Kept: ${keptTagName ?? 'none'}`,
      `${dryRun ? 'Would delete' : 'Deleted'}: ${prereleaseDeleted} prerelease(s)`,
      `${dryRun ? 'Would delete' : 'Deleted'} orphan tags: ${orphanDeleted}`,
      `Failed: ${failed}`,
    ])
    .write();
}

module.exports = async ({ github, context, core, dryRun = true }) => {
  const { owner, repo } = context.repo;

  const prereleases = await fetchPrereleasesNewestFirst({ github, owner, repo });
  const keptTagName = prereleases[0]?.tag_name ?? null;
  const attachedTagNames = new Set(prereleases.map(release => release.tag_name));

  if (keptTagName) {
    console.log(`Keeping latest prerelease: ${keptTagName}`);
  }

  const prereleaseResult = await deleteOldPrereleases({ github, core, owner, repo, prereleases, dryRun });
  const orphanResult = await sweepOrphanTags({ github, core, owner, repo, dryRun, attachedTagNames });

  const failed = prereleaseResult.failed + orphanResult.failed;

  await writeSummary({
    core,
    dryRun,
    keptTagName,
    prereleaseDeleted: prereleaseResult.deleted,
    orphanDeleted: orphanResult.deleted,
    failed,
  });

  if (failed > 0) {
    core.setFailed(`${failed} delete operation(s) failed`);
  }

  return {
    kept: keptTagName,
    deleted: prereleaseResult.deleted,
    orphanDeleted: orphanResult.deleted,
    failed,
    dryRun,
  };
};
