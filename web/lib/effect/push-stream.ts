import { type Cause, Effect, Queue, Stream } from "effect";

export type PushStream<A> = {
  /** The values pushed so far and from now on, until `end`. Consume it once. */
  stream: Stream.Stream<A>;
  /** Adds a value to the stream; ignored once the stream has ended. */
  push: (value: A) => void;
  /** Ends the stream after the values already pushed. */
  end: () => void;
};

/**
 * Returns a stream fed by `push`, such as the requests of a client-streaming or bidirectional
 * call fed by events. Values pushed before the stream is consumed are kept, without a limit, so
 * nothing is dropped while a call starts or waits for the server to read.
 */
export function createPushStream<A>(): PushStream<A> {
  const queue = Effect.runSync(Queue.unbounded<A, Cause.Done>());
  return {
    stream: Stream.fromQueue(queue),
    push: (value) => {
      Queue.offerUnsafe(queue, value);
    },
    end: () => {
      Queue.endUnsafe(queue);
    },
  };
}
