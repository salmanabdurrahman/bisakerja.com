"use client";

import { useMemo, useState } from "react";
import { Button } from "@/components/ui/button";
import { APIRequestError } from "@/lib/utils/fetch-json";
import { createSessionAPIClient } from "@/services/session-api-client";

interface BookmarkButtonProps {
  jobID: string;
  initialIsBookmarked: boolean;
}

export function BookmarkButton({
  jobID,
  initialIsBookmarked,
}: BookmarkButtonProps) {
  const sessionClient = useMemo(() => createSessionAPIClient(), []);
  const [isBookmarked, setIsBookmarked] = useState(initialIsBookmarked);
  const [isPending, setIsPending] = useState(false);

  async function handleToggle() {
    setIsPending(true);
    // optimistic update
    setIsBookmarked((prev) => !prev);
    try {
      if (!isBookmarked) {
        await sessionClient.createBookmark(jobID);
      } else {
        await sessionClient.deleteBookmark(jobID);
      }
    } catch (error) {
      // revert on failure
      setIsBookmarked((prev) => !prev);
      if (!(error instanceof APIRequestError)) throw error;
    } finally {
      setIsPending(false);
    }
  }

  return (
    <Button
      type="button"
      variant="outline"
      size="sm"
      onClick={() => void handleToggle()}
      disabled={isPending}
      aria-label={isBookmarked ? "Remove bookmark" : "Bookmark job"}
    >
      {isBookmarked ? "★ Bookmarked" : "☆ Bookmark"}
    </Button>
  );
}
