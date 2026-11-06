import test, { expect } from '@playwright/test';
import { cleanupBackend } from '../utils/cleanup.util';
import { unsavedChangesBar } from '../utils/unsaved-changes.util';

test.beforeEach(async () => await cleanupBackend());

test('Ctrl+A toggles the account menu without replacing text selection', async ({ page }) => {
	await page.goto('/settings/account');
	const displayName = page.getByLabel('Display Name');
	await expect(displayName).toBeVisible();
	await displayName.evaluate((el: HTMLInputElement) => el.blur());

	await page.keyboard.press('Control+A');
	await expect(page.getByRole('menu')).toBeVisible();
	await page.keyboard.press('Control+A');
	await expect(page.getByRole('menu')).not.toBeVisible();

	await displayName.fill('Keyboard shortcut');
	await page.keyboard.press('Control+A');
	await expect(page.getByRole('menu')).not.toBeVisible();
	expect(
		await displayName.evaluate((el: HTMLInputElement) =>
			el.value.slice(el.selectionStart ?? 0, el.selectionEnd ?? 0)
		)
	).toBe('Keyboard shortcut');
});

test('Ctrl+B toggles the sidebar on desktop', async ({ page }) => {
	await page.goto('/settings/account');
	await expect(page.getByLabel('Display Name')).toBeVisible();
	await page.getByLabel('Display Name').evaluate((el: HTMLInputElement) => el.blur());

	const collapse = page.getByRole('button', { name: 'Collapse sidebar' });
	const expand = page.getByRole('button', { name: 'Expand sidebar' });
	const startsCollapsed = await expand.isVisible();

	await page.keyboard.press('Control+B');
	await expect(startsCollapsed ? collapse : expand).toBeVisible();
	await page.keyboard.press('Control+B');
	await expect(startsCollapsed ? expand : collapse).toBeVisible();
});

test('Ctrl+Enter saves and Ctrl+Backspace discards pending changes', async ({ page }) => {
	await page.goto('/settings/account');
	const displayName = page.getByLabel('Display Name');
	await expect(displayName).toBeVisible();
	const original = await displayName.inputValue();

	await displayName.fill('Shortcut save');
	await expect(unsavedChangesBar(page)).toBeVisible();
	await page.keyboard.press('Control+Enter');
	await expect(unsavedChangesBar(page).getByText('Changes saved', { exact: true })).toBeVisible();
	const saved = await displayName.inputValue();
	expect(saved).not.toBe(original);

	await displayName.fill('Shortcut discard');
	await expect(unsavedChangesBar(page)).toBeVisible();
	await displayName.evaluate((el: HTMLInputElement) => el.blur());
	await page.keyboard.press('Control+Backspace');
	await expect(unsavedChangesBar(page)).toHaveCount(0);
	await expect(displayName).toHaveValue(saved);
});
