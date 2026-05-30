export function durationHours(timeArrived: Date, timeLeft: Date): number {
  const ms = timeLeft.getTime() - timeArrived.getTime();
  return ms / (1000 * 60 * 60);
}

export function sumHours(
  entries: { timeArrived: Date; timeLeft: Date }[],
): number {
  return entries.reduce(
    (total, e) => total + durationHours(e.timeArrived, e.timeLeft),
    0,
  );
}
