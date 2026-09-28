type PlayerTrack = {
  attachmentId: string;
  kind: "audio" | "video";
  fileName: string;
  messageId: string;
  parentId: string | undefined;
  channelId: string;
  workspaceId: string;
  authorName: string;
  // 計測できなかったときは再生を始めてからメディアの値で埋める
  durationSeconds: number;
};

type PlayerState = {
  track: PlayerTrack | null;
  isPlaying: boolean;
  position: number;
  duration: number;
  rate: number;
  // 再生中の添付がメッセージ一覧に見えているか。見えていなければミニプレイヤーを出す
  isInlineVisible: boolean;
};

type SlotKind = "inline" | "mini";

const RATES = [1, 1.5, 2];

const initialState: PlayerState = {
  duration: 0,
  isInlineVisible: false,
  isPlaying: false,
  position: 0,
  rate: 1,
  track: null,
};

// アプリ全体で <video> と <audio> を 1 つずつだけ持ち、画面の切り替えで再生が止まらないようにする
export const createMediaPlayer = () => {
  let state = initialState;
  const listeners = new Set<() => void>();
  const slots: Record<SlotKind, HTMLElement[]> = { inline: [], mini: [] };
  let inlineElement: HTMLElement | null = null;
  let observer: IntersectionObserver | null = null;
  let resolveUrl: (() => Promise<string>) | null = null;
  let elements: { audio: HTMLAudioElement; video: HTMLVideoElement; holder: HTMLElement } | null =
    null;

  const setState = (patch: Partial<PlayerState>) => {
    state = { ...state, ...patch };
    for (const listener of listeners) {
      listener();
    }
  };

  const bind = <T extends HTMLMediaElement>(element: T) => {
    element.addEventListener("timeupdate", () => {
      setState({ position: element.currentTime });
    });
    element.addEventListener("durationchange", () => {
      if (Number.isFinite(element.duration)) {
        setState({ duration: element.duration });
      }
    });
    element.addEventListener("play", () => {
      setState({ isPlaying: true });
    });
    element.addEventListener("pause", () => {
      setState({ isPlaying: false });
    });
    element.addEventListener("ended", () => {
      setState({ isPlaying: false, position: 0 });
    });
    // 署名付き URL の期限が切れたら取り直し、同じ位置から続ける
    element.addEventListener("error", () => {
      const retry = resolveUrl;
      resolveUrl = null;
      if (retry === null || state.track === null) {
        setState({ isPlaying: false });
        return;
      }
      const { position } = state;
      void retry().then((url) => {
        element.src = url;
        element.currentTime = position;
        return element.play();
      });
    });
    return element;
  };

  const createElements = () => {
    // どのスロットにもないときの置き場所。DOM から外すと再生が止まるため
    const holder = document.createElement("div");
    holder.setAttribute("aria-hidden", "true");
    holder.style.cssText = "position:fixed;width:0;height:0;overflow:hidden;";
    document.body.append(holder);
    const video = bind(document.createElement("video"));
    video.playsInline = true;
    video.className = "block size-full object-contain";
    return { audio: bind(document.createElement("audio")), holder, video };
  };

  const getElements = () => (elements ??= createElements());

  const current = () => {
    if (state.track === null || elements === null) {
      return null;
    }
    return state.track.kind === "video" ? elements.video : elements.audio;
  };

  // 見えているメッセージの枠を優先し、なければミニプレイヤーに映す
  const place = () => {
    if (state.track?.kind !== "video" || elements === null) {
      return;
    }
    const target =
      (state.isInlineVisible ? slots.inline.at(-1) : undefined) ??
      slots.mini.at(-1) ??
      slots.inline.at(-1) ??
      elements.holder;
    if (elements.video.parentElement !== target) {
      target.append(elements.video);
    }
  };

  const setInlineVisible = (isInlineVisible: boolean) => {
    if (state.isInlineVisible !== isInlineVisible) {
      setState({ isInlineVisible });
      place();
    }
  };

  const stop = () => {
    const element = current();
    if (element !== null) {
      element.pause();
      element.removeAttribute("src");
      element.load();
    }
    resolveUrl = null;
    setState({ duration: 0, isPlaying: false, position: 0, track: null });
  };

  const play = async (track: PlayerTrack, getUrl: () => Promise<string>) => {
    stop();
    const { audio, video } = getElements();
    const element = track.kind === "video" ? video : audio;
    setState({ duration: track.durationSeconds, isPlaying: true, track });
    place();
    resolveUrl = getUrl;
    try {
      element.src = await getUrl();
      if (state.track !== track) {
        return;
      }
      element.playbackRate = state.rate;
      await element.play();
    } catch (error) {
      if (state.track === track) {
        stop();
      }
      throw error;
    }
  };

  return {
    // メッセージ内の再生中の添付を監視し、画面外に出たらミニプレイヤーへ切り替える
    attachInline: (element: HTMLElement) => {
      inlineElement = element;
      observer?.disconnect();
      observer = new IntersectionObserver(([entry]) => {
        setInlineVisible(entry?.isIntersecting ?? false);
      });
      observer.observe(element);
      return () => {
        if (inlineElement === element) {
          observer?.disconnect();
          observer = null;
          inlineElement = null;
          setInlineVisible(false);
        }
      };
    },
    cycleRate: () => {
      const rate = RATES[(RATES.indexOf(state.rate) + 1) % RATES.length] ?? 1;
      const element = current();
      if (element !== null) {
        element.playbackRate = rate;
      }
      setState({ rate });
    },
    getState: () => state,
    pause: () => {
      current()?.pause();
    },
    play,
    registerSlot: (kind: SlotKind, element: HTMLElement) => {
      slots[kind].push(element);
      place();
      return () => {
        slots[kind] = slots[kind].filter((slot) => slot !== element);
        place();
      };
    },
    resume: async () => {
      await current()?.play();
    },
    // 画面内に元のメッセージがあればスクロールする。なければ false を返し、呼び出し側で遷移する
    scrollToSource: () => {
      if (inlineElement === null) {
        return false;
      }
      inlineElement.scrollIntoView({ behavior: "smooth", block: "center" });
      return true;
    },
    seek: (position: number) => {
      const element = current();
      if (element !== null) {
        element.currentTime = position;
      }
      setState({ position });
    },
    stop,
    subscribe: (listener: () => void) => {
      listeners.add(listener);
      return () => {
        listeners.delete(listener);
      };
    },
  };
};

export const mediaPlayer = createMediaPlayer();
