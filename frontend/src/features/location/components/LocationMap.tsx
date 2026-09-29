import "leaflet/dist/leaflet.css";
import { Circle, CircleMarker, MapContainer, TileLayer } from "react-leaflet";

import { cn } from "#/components/ui/styles/styles";

import type { MessageLocation } from "#/gen/chat/v1/message_pb";

import type { LatLngTuple } from "leaflet";

type LocationMapProps = {
  location: MessageLocation;
  className: string;
};

const attribution =
  '&copy; <a href="https://www.openstreetmap.org/copyright" target="_blank" rel="noopener noreferrer">OpenStreetMap</a>';

// 誤差の円が収まるくらいに寄せる
const zoomFor = (accuracyMeters: number | undefined) =>
  accuracyMeters === undefined
    ? 16
    : Math.max(11, Math.min(17, Math.round(19 - Math.log2(Math.max(accuracyMeters, 1)))));

// OpenStreetMap のタイルで地点と測位の誤差を描く。Leaflet の z-index が周りに漏れないよう isolate する
export const LocationMap = ({ location, className }: LocationMapProps) => {
  const center: LatLngTuple = [location.latitude, location.longitude];
  return (
    <MapContainer
      key={center.join(",")}
      center={center}
      zoom={zoomFor(location.accuracyMeters)}
      scrollWheelZoom={false}
      className={cn("isolate bg-sunken", className)}
    >
      <TileLayer url="https://tile.openstreetmap.org/{z}/{x}/{y}.png" attribution={attribution} />
      {location.accuracyMeters !== undefined && (
        <Circle
          center={center}
          radius={location.accuracyMeters}
          pathOptions={{ className: "fill-accent stroke-accent", fillOpacity: 0.15, weight: 1 }}
        />
      )}
      <CircleMarker
        center={center}
        radius={7}
        pathOptions={{ className: "fill-accent stroke-surface", fillOpacity: 1, weight: 2 }}
      />
    </MapContainer>
  );
};
