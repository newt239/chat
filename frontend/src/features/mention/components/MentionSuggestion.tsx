import { Avatar, Text, Loader } from "@mantine/core";

import type { MentionSuggestion } from "../types";

type MentionSuggestionProps = {
  suggestions: MentionSuggestion[];
  isLoading: boolean;
  onSelect: (suggestion: MentionSuggestion) => void;
  selectedIndex?: number;
}

export const MentionSuggestionList = ({
  suggestions,
  isLoading,
  onSelect,
  selectedIndex = -1,
}: MentionSuggestionProps) => {
  if (isLoading) {
    return (
      <div className="absolute z-50 bg-white border border-gray-200 rounded-md shadow-lg p-2 min-w-48">
        <div className="flex items-center space-x-2 p-2">
          <Loader size="xs" />
          <Text size="sm" c="dimmed">
            Searching...
          </Text>
        </div>
      </div>
    );
  }

  if (suggestions.length === 0) {
    return null;
  }

  return (
    <div className="absolute z-50 bg-white border border-gray-200 rounded-md shadow-lg p-2 min-w-48 max-h-60 overflow-y-auto">
      {suggestions.map((suggestion, index) => (
        <button
          key={`${suggestion.type}-${suggestion.id}`}
          onClick={() => onSelect(suggestion)}
          className={`w-full flex items-center space-x-2 p-2 rounded hover:bg-gray-100 ${
            index === selectedIndex ? "bg-blue-50" : ""
          }`}
          type="button"
        >
          <Avatar src={suggestion.avatarUrl} alt={suggestion.name} size="sm" radius="xl">
            {suggestion.name.charAt(0).toUpperCase()}
          </Avatar>

          <div className="flex-1 text-left">
            <Text size="sm" fw={500}>
              {suggestion.name}
            </Text>
            <Text size="xs" c="dimmed">
              {suggestion.type === "user" ? "User" : "Group"}
            </Text>
          </div>

          <Text size="xs" c="dimmed">
            @
          </Text>
        </button>
      ))}
    </div>
  );
};
