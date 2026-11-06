import { z } from 'zod/v4';

export const groupNameSchema = z.object({
	friendlyName: z.string().min(2).max(50),
	name: z.string().min(2).max(255),
	claimPriority: z.coerce.number().int().min(0).max(1000)
});
