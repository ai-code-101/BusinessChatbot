import ChatWidget from "./components/ChatWidget.jsx";

export default function App() {
  return (
    <div style={{ height: "100%" }}>
      <div className="demo-page">
        This is a stand-in for your business's actual website.
        <br />
        The chat widget in the bottom-right corner is what gets embedded
        on the real site — this page is only here for testing it.
      </div>
      <ChatWidget
        businessId={import.meta.env.VITE_BUSINESS_ID || "default_business"}
        businessName={import.meta.env.VITE_BUSINESS_NAME || "Peak Mobile"}
      />
    </div>
  );
}
